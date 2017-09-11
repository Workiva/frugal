package examples;

import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.server.FDefaultNettyHttpProcessor;
import com.workiva.frugal.server.FNettyHttpHandler;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import v1.music.FStore;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

/**
 * Creates an HTTP server listening for incoming requests.
 */
public class HttpServer {

    public static void main(String[] args) throws InterruptedException, IOException, TimeoutException, TException {

        // Specify the protocol used for serializing requests.
        // Clients must use the same protocol stack
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        // Create a new server processor.
        // Incoming requests to the server are passed to the processor.
        // Results from the processor are returned back to the client.
        FStore.Processor processor = new FStore.Processor(new FStoreHandler(), new LoggingMiddleware());

        // Configure the server.
        EventLoopGroup bossGroup = new NioEventLoopGroup(1);
        EventLoopGroup workerGroup = new NioEventLoopGroup();
        FNettyHttpHandler httpHandler = FNettyHttpHandler.of(FDefaultNettyHttpProcessor.of(processor, protocolFactory));
        try {
            ServerBootstrap b = new ServerBootstrap();
            b.group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class).option(ChannelOption.SO_BACKLOG, 100)
                    .handler(new LoggingHandler(LogLevel.INFO))
                    .childHandler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        public void initChannel(SocketChannel ch) throws Exception {
                            ChannelPipeline p = ch.pipeline();
                            ch.pipeline().addLast("codec", new HttpServerCodec());
                            ch.pipeline().addLast("aggregator", new HttpObjectAggregator(512*1024));
                            p.addLast(httpHandler);
                        }
                    });
            // Start the server.
            ChannelFuture f = b.bind(9090).sync();
            // Wait until the server socket is closed.
            f.channel().closeFuture().sync();
        } finally {
            // Shut down all event loops to terminate all threads.
            bossGroup.shutdownGracefully();
            workerGroup.shutdownGracefully();
        }
    }
}
