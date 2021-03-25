package thrift;

import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpRequestDecoder;
import io.netty.handler.codec.http.HttpResponseEncoder;
import io.netty.handler.codec.http.HttpUtil;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TException;
import org.apache.thrift.TProcessor;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TCompactProtocol;
import org.apache.thrift.protocol.TJSONProtocol;
import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.server.TServer;
import org.apache.thrift.server.TSimpleServer;
import org.apache.thrift.transport.TMemoryBuffer;
import org.apache.thrift.transport.TServerSocket;
import org.apache.thrift.transport.TServerTransport;
import org.apache.thrift.transport.TSocket;
import org.apache.thrift.transport.TTransport;
import org.junit.Assert;
import org.junit.Test;

import java.nio.ByteBuffer;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

import static io.netty.handler.codec.http.HttpHeaderNames.CONNECTION;

public class ThriftJsonTest {

    @Test
    public void testNewTProtocolFactoryCompact() throws TException {
        TProtocolFactory tpf = new TCompactProtocol.Factory();
        testLargeFileTransport(tpf);
    }

    @Test
    public void testNewTProtocolFactoryBinary() throws TException {
        TProtocolFactory tpf = new TBinaryProtocol.Factory();
        testLargeFileTransport(tpf);

    }

    @Test
    public void testNewTProtocolFactoryJson() throws TException {
        TProtocolFactory tpf = new TJSONProtocol.Factory();
        testLargeFileTransport(tpf);
    }

    public void testLargeFileTransport(TProtocolFactory tpf) throws TException {
        System.out.println("Dan was here");
        int big = 100;

        TMemoryBuffer transport = new TMemoryBuffer(big);
        TProtocol inProtocol = tpf.getProtocol(transport);

        byte[] message = new byte[big];
        String expected = "yeeeeeeeeeeeehaw";
        inProtocol.writeString(expected);

        TProtocol outProtocol = tpf.getProtocol(transport);
        String str = outProtocol.readString();
        Assert.assertEquals(str, expected);
    }

    public static TestThriftService.Iface handler;

    public static TestThriftService.Processor processor;


    class TestThriftServiceHandler implements TestThriftService.Iface {

        @Override
        public String testString(String s) throws TException {
            System.out.println(s);
            return s;
        }


    }

    public Thread runServer() {
        try {
            handler = new ThriftJsonTest.TestThriftServiceHandler();
            processor = new TestThriftService.Processor(handler);

            Runnable simple = new Runnable() {
                public void run() {
                    simple(processor);
                }
            };

            Thread thread = new Thread(simple);
            thread.start();
            return thread;
        } catch (Exception x) {
            x.printStackTrace();
            throw x;
        }
    }

    public static void simple(TestThriftService.Processor processor) {
        try {
            TServerTransport serverTransport = new TServerSocket(9090);
            TServer server = new TSimpleServer(new TServer.Args(serverTransport).processor(processor).protocolFactory(new TJSONProtocol.Factory()));

            System.out.println("Starting the simple server...");
            server.serve();
            System.out.println("please no");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    public void runClient() {

        try {
            TTransport transport;

            transport = new TSocket("localhost", 9090);
            transport.open();

            TProtocol protocol = new TJSONProtocol(transport);
            TestThriftService.Client client = new TestThriftService.Client(protocol);

            perform(client);

            transport.close();
        } catch (TException x) {
            x.printStackTrace();
        }
    }

    private void perform(TestThriftService.Client client) throws TException
    {

        String str = client.testString("yeehaw");
        System.out.println("returned: " + str);
    }

    @Test
    public void testSimpleThriftExample() throws Exception {
        runServer();
        runClient();
    }
}
