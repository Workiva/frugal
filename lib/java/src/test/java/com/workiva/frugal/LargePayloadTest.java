package com.workiva.frugal;

import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.provider.FServiceProvider;
import com.workiva.frugal.server.FDefaultNettyHttpProcessor;
import com.workiva.frugal.server.FNettyHttpHandler;
import com.workiva.frugal.server.FNettyHttpProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FHttpTransport;
import com.workiva.frugal.transport.FTransport;
import frugal.test.Insanity;
import frugal.test.Numberz;
import frugal.test.Xception;
import frugal.test.Xception2;
import frugal.test.Xtruct;
import frugal.test.Xtruct2;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpRequestDecoder;
import io.netty.handler.codec.http.HttpResponseEncoder;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TCompactProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.junit.Test;
import frugal.test.FFrugalTest;

import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

public class LargePayloadTest {

  public LargePayloadTest() {
  }

  public class FrugalTestHandler implements FFrugalTest.Iface {

    // Each RPC handler "test___" accepts a value of type ___ and returns the same value (where applicable).
    // The client then asserts that the returned value is equal to the value sent.
    @Override
    public void testVoid(FContext ctx) throws TException {
    }

    @Override
    public String testString(FContext ctx, String thing) throws TException {
      return thing;
    }

    @Override
    public boolean testBool(FContext ctx, boolean thing) throws TException {
      return thing;
    }

    @Override
    public byte testByte(FContext ctx, byte thing) throws TException {
      return thing;
    }

    @Override
    public int testI32(FContext ctx, int thing) throws TException {
      return thing;
    }

    @Override
    public long testI64(FContext ctx, long thing) throws TException {
      return thing;
    }

    @Override
    public double testDouble(FContext ctx, double thing) throws TException {
      return thing;
    }

    @Override
    public ByteBuffer testBinary(FContext ctx, ByteBuffer thing) throws TException {
      return thing;
    }

    @Override
    public Xtruct testStruct(FContext ctx, Xtruct thing) throws TException {
      return thing;
    }

    @Override
    public Xtruct2 testNest(FContext ctx, Xtruct2 thing) throws TException {
      return thing;
    }

    @Override
    public Map<Integer, Integer> testMap(FContext ctx, Map<Integer, Integer> thing) throws TException {
      return thing;
    }

    @Override
    public Map<String, String> testStringMap(FContext ctx, Map<String, String> thing) throws TException {
      return thing;
    }

    @Override
    public Set<Integer> testSet(FContext ctx, Set<Integer> thing) throws TException {
      return thing;
    }

    @Override
    public List<Integer> testList(FContext ctx, List<Integer> thing) throws TException {
      return thing;
    }

    @Override
    public Numberz testEnum(FContext ctx, Numberz thing) throws TException {
      return thing;
    }

    @Override
    public long testTypedef(FContext ctx, long thing) throws TException {
      return thing;
    }

    @Override
    public Map<Integer, Map<Integer, Integer>> testMapMap(FContext ctx, int hello) throws TException {
      Map<Integer, Integer> mp1 = new HashMap<>();
      mp1.put(-4,-4);
      mp1.put(-3,-3);
      mp1.put(-2,-2);
      mp1.put(-1,-1);

      Map<Integer, Integer> mp2 = new HashMap<>();
      mp2.put(4,4);
      mp2.put(3,3);
      mp2.put(2,2);
      mp2.put(1,1);

      Map<Integer, Map<Integer, Integer>> rMapMap = new HashMap<>();
      rMapMap.put(-4, mp1);
      rMapMap.put(4, mp2);
      return rMapMap;
    }

    @Override
    public boolean TestUppercaseMethod(FContext ctx, boolean thing) throws TException {
      return thing;
    }

    @Override
    public Map<Long, Map<Numberz, Insanity>> testInsanity(FContext ctx, Insanity argument) throws TException {
      Map<Numberz, Insanity> mp1 = new HashMap<>();
      mp1.put(Numberz.findByValue(2), argument);
      mp1.put(Numberz.findByValue(3), argument);

      Map<Numberz, Insanity> mp2 = new HashMap<>();

      Map<Long, Map<Numberz, Insanity>> returnInsanity = new HashMap<>();
      returnInsanity.put((long) 1, mp1);
      returnInsanity.put((long) 2, mp2);

      return returnInsanity;
    }

    @Override
    public Xtruct testMulti(FContext ctx, byte arg0, int arg1, long arg2, Map<Short, String> arg3, Numberz arg4, long arg5) throws TException {
      Xtruct r = new Xtruct();

      r.string_thing = "Hello2";
      r.byte_thing = arg0;
      r.i32_thing = arg1;
      r.i64_thing = arg2;

      return r;
    }

    @Override
    public void testException(FContext ctx, String arg) throws TException {
      switch (arg) {
        case "Xception":
          Xception e = new Xception();
          e.errorCode = 1001;
          e.message = arg;
          throw e;
        case "TException":
          throw new TException("Just TException");
        default:
      }
    }

    // This doesn't really make the same sense to check in Java
    // as in other languages, because we rethrow any caught runtime exceptions.
    @Override
    public void testUncaughtException(FContext ctx) throws TException {
      throw new TException("An uncaught error which will be caught in Java");
    }

    @Override
    public void testUncheckedTApplicationException(FContext ctx) throws TException {
      throw new TApplicationException(400, "Unchecked TApplicationException");
    }

    @Override
    public Xtruct testMultiException(FContext ctx, String arg0, String arg1) throws TException {
      switch (arg0) {
        case "Xception":
          Xception e = new Xception();
          e.errorCode = 1001;
          e.message = "This is an Xception";
          throw e;
        case "Xception2":
          Xception2 e2 = new Xception2();
          e2.errorCode = 2002;
          e2.struct_thing = new Xtruct();
          e2.struct_thing.string_thing = "This is an Xception2";
          throw e2;
        default:
          Xtruct r = new Xtruct();
          r.string_thing = arg1;
          return r;
      }
    }

    @Override
    public void testRequestTooLarge(FContext ctx, java.nio.ByteBuffer request) throws TException {
      throw new TApplicationException(400, "testRequestTooLarge should" +
          " never be sucessfully called.");
    }

    @Override
    public java.nio.ByteBuffer testResponseTooLarge(FContext ctx, java.nio.ByteBuffer request) {
      java.nio.ByteBuffer response = ByteBuffer.allocate(1024*1024);
      return response;
    }

    @Override
    public void testOneway(FContext ctx, int secondsToSleep) throws TException {
    }

    @Override
    public void testSuperClass(FContext ctx) throws TException {
    }

  }

  public static class FNettyHttpHandlerFactory {

    final FProcessor processor;
    final FProtocolFactory protocolFactory;

    FNettyHttpHandlerFactory(FProcessor processor, FProtocolFactory protocolFactory) {
      this.processor = processor;
      this.protocolFactory = protocolFactory;
    }

    public FNettyHttpHandler newHandler() {
      FNettyHttpProcessor httpProcessor = FDefaultNettyHttpProcessor.of(processor, protocolFactory);
      return FNettyHttpHandler.of(httpProcessor);
    }

  }

  public static class NettyServerThread extends Thread {
    Integer port;
    final FNettyHttpHandlerFactory handlerFactory;

    NettyServerThread(Integer port, FNettyHttpHandlerFactory handlerFactory) {
      this.port = port;
      this.handlerFactory = handlerFactory;
    }

    public void run() {
      EventLoopGroup bossGroup = new NioEventLoopGroup(1);
      EventLoopGroup workerGroup = new NioEventLoopGroup();
      try {
        ServerBootstrap b = new ServerBootstrap();
        b.group(bossGroup, workerGroup)
            .channel(NioServerSocketChannel.class)
            .handler(new LoggingHandler(LogLevel.INFO))
            .childHandler(new LargePayloadTest.NettyHttpInitializer(handlerFactory));

        Channel ch = b.bind(port).sync().channel();

        ch.closeFuture().sync();
      } catch (InterruptedException e) {
        e.printStackTrace();
      } finally {
        bossGroup.shutdownGracefully();
        workerGroup.shutdownGracefully();
      }
    }
  }

  public static class NettyHttpInitializer extends ChannelInitializer<SocketChannel> {

    LargePayloadTest.FNettyHttpHandlerFactory handlerFactory;

    public NettyHttpInitializer(LargePayloadTest.FNettyHttpHandlerFactory handlerFactory) {
      this.handlerFactory = handlerFactory;
    }

    @Override
    public void initChannel(SocketChannel ch) {
      ChannelPipeline p = ch.pipeline();
      p.addLast(new HttpRequestDecoder());
      p.addLast(new HttpObjectAggregator(500 * 1024 * 1024));
      p.addLast(new HttpResponseEncoder());
      p.addLast(handlerFactory.newHandler());
    }
  }

  private static class ServerMiddleware implements ServiceMiddleware {
    CountDownLatch called;

    ServerMiddleware(CountDownLatch called) {
      this.called = called;
    }

    @Override
    public <T> InvocationHandler<T> apply(T next) {
      return new InvocationHandler<T>(next) {
        @Override
        public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
          Object[] subArgs = Arrays.copyOfRange(args, 1, args.length);
//          System.out.printf("%s(%s)\n", method.getName(), Arrays.toString(subArgs));
          if (method.getName().equals("testOneway")) {

            called.countDown();
          }
          return method.invoke(receiver, args);
        }
      };
    }
  }
  public static boolean middlewareCalled = false;
  public static class ClientMiddleware implements ServiceMiddleware {

    @Override
    public <T> InvocationHandler<T> apply(T next) {
      return new InvocationHandler<T>(next) {
        @Override
        public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
          Object[] subArgs = Arrays.copyOfRange(args, 1, args.length);
//          System.out.printf("%s", method.getName());
//          System.out.printf("%s(%s) = ", method.getName(), Arrays.toString(subArgs));
          middlewareCalled = true;
          try {
            Object ret = method.invoke(receiver, args);
            System.out.printf("%s \n", ret);
            return ret;
          } catch (Exception e) {
            throw e;
          }
        }
      };
    }
  }


  @Test
  public void endToEnd() throws TException {
//    TProtocolFactory tpf = new TJSONProtocol.Factory();
    TProtocolFactory tpf = new TCompactProtocol.Factory();
    FProtocolFactory fProtocolFactory = new FProtocolFactory(tpf);
    CountDownLatch called = new CountDownLatch(1);

    FFrugalTest.Iface handler = new FrugalTestHandler();
    FFrugalTest.Processor processor = new FFrugalTest.Processor(handler, new ServerMiddleware(called));

    FNettyHttpHandlerFactory handlerFactory = new FNettyHttpHandlerFactory(processor, fProtocolFactory);
    NettyServerThread serverThread = new NettyServerThread(8080, handlerFactory);
    serverThread.start();

    String url = "http://localhost:8080";
//    CloseableHttpClient httpClient = HttpClients.createDefault();
    CloseableHttpClient httpClient = HttpClients.custom().setConnectionTimeToLive(60, TimeUnit.SECONDS).build();

    // Set request and response size limit to 1mb
    int maxSize = 250 * 1024 * 1024;
    FHttpTransport.Builder httpTransport = new FHttpTransport.Builder(httpClient, url).withRequestSizeLimit(maxSize).withResponseSizeLimit(maxSize);
    FTransport fTransport = httpTransport.build();
    fTransport.open();

    FFrugalTest.Client testClient = new FFrugalTest.Client(new FServiceProvider(fTransport, fProtocolFactory, new ClientMiddleware()));
    FContext context = new FContext("context");
    byte[] data = new byte[150 * 1024 * 1024];

    ByteBuffer bin = testClient.testBinary(context, ByteBuffer.wrap(data));  }
}

