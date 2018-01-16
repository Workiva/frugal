package com.workiva.frugal.server;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import io.nats.client.AsyncSubscription;
import io.nats.client.Connection;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TJSONProtocol;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import org.mockito.ArgumentCaptor;

import java.io.IOException;
import java.util.Arrays;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.ThreadPoolExecutor;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.verifyNoMoreInteractions;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FNatsServer}.
 */
@RunWith(JUnit4.class)
public class FNatsServerTest {

    private Connection mockConn;
    private FProcessor mockProcessor;
    private FProtocolFactory mockProtocolFactory;
    private String subject = "foo";
    private String queue = "bar";
    private FNatsServer server;

    @Before
    public void setUp() {
        mockConn = mock(Connection.class);
        mockProcessor = mock(FProcessor.class);
        mockProtocolFactory = mock(FProtocolFactory.class);
        server = new FNatsServer.Builder(mockConn, mockProcessor, mockProtocolFactory, new String[]{subject})
                .withQueueGroup(queue).build();
    }

    @Test
    public void testBuilderConfiguresServer() {
        FNatsServer server =
                new FNatsServer.Builder(mockConn, mockProcessor, mockProtocolFactory, new String[]{subject})
                        .withHighWatermark(7)
                        .withQueueGroup("myQueue")
                        .withQueueLength(7)
                        .withWorkerCount(10)
                        .build();

        assertEquals(server.getQueue(), "myQueue");
        assertEquals(((ThreadPoolExecutor) server.getExecutorService()).getQueue().remainingCapacity(), 7);
        assertEquals(((ThreadPoolExecutor) server.getExecutorService()).getMaximumPoolSize(), 10);
    }

    @Test
    public void testServe() throws TException, IOException, InterruptedException {
        server = new FNatsServer.Builder(mockConn, mockProcessor, mockProtocolFactory, new String[]{subject})
                .withQueueGroup(queue).build();
        ArgumentCaptor<String> subjectCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> queueCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(mockConn.subscribe(subjectCaptor.capture(), queueCaptor.capture(), handlerCaptor.capture()))
                .thenReturn(sub);

        CountDownLatch stopSignal = new CountDownLatch(1);

        // start/stop the server
        new Thread(() -> {
            try {
                server.serve();
                stopSignal.countDown(); // signal server stop
            } catch (TException e) {
                fail(e.getMessage());
            }
        }).start();
        server.stop();

        stopSignal.await(); // wait for orderly shutdown

        assertEquals(subject, subjectCaptor.getValue());
        assertEquals(queue, queueCaptor.getValue());
        assertNotNull(handlerCaptor.getValue());
        verify(sub).unsubscribe();
    }

    @Test
    public void testRequestHandler() {
        ExecutorService executor = mock(ExecutorService.class);
        FNatsServer server =
                new FNatsServer.Builder(mockConn, mockProcessor, mockProtocolFactory, new String[]{subject})
                        .withExecutorService(executor).build();
        MessageHandler handler = server.newRequestHandler();
        String reply = "reply";
        byte[] data = "this is a request".getBytes();
        Message msg = new Message(subject, reply, data);

        handler.onMessage(msg);

        ArgumentCaptor<Runnable> captor = ArgumentCaptor.forClass(Runnable.class);
        verify(executor).execute(captor.capture());
        assertEquals(FNatsServer.Request.class, captor.getValue().getClass());
        FNatsServer.Request request = (FNatsServer.Request) captor.getValue();
        assertArrayEquals(data, request.frameBytes);
        assertEquals(reply, request.reply);
        assertEquals(5000, request.highWatermark);
        assertEquals(mockProtocolFactory, request.inputProtoFactory);
        assertEquals(mockProtocolFactory, request.outputProtoFactory);
        assertEquals(mockProcessor, request.processor);
        assertEquals(mockConn, request.conn);
    }

    @Test
    public void testRequestHandler_noReply() {
        ExecutorService executor = mock(ExecutorService.class);
        FNatsServer server =
                new FNatsServer.Builder(mockConn, mockProcessor, mockProtocolFactory, new String[]{subject})
                .withExecutorService(executor).build();
        MessageHandler handler = server.newRequestHandler();
        byte[] data = "this is a request".getBytes();
        Message msg = new Message(subject, null, data);

        handler.onMessage(msg);

        verifyNoMoreInteractions(executor);
    }

    @Test
    public void testRequestProcess() throws TException, IOException {
        byte[] data = "xxxxhello".getBytes();
        long timestamp = System.currentTimeMillis();
        String reply = "reply";
        long highWatermark = 5000;
        MockFProcessor processor = new MockFProcessor(data, "blah".getBytes());
        mockProtocolFactory = new FProtocolFactory(new TJSONProtocol.Factory());
        FNatsServer.Request request = new FNatsServer.Request(data, timestamp, reply, highWatermark,
                mockProtocolFactory, mockProtocolFactory, processor, mockConn);

        request.run();

        byte[] expected = new byte[]{0, 0, 0, 6, 34, 98, 108, 97, 104, 34};
        verify(mockConn).publish(reply, expected);
    }

    @Test
    public void testRequestProcessRuntimeException() throws TException, IOException {
        byte[] data = "xxxxhello".getBytes();
        long timestamp = System.currentTimeMillis();
        String reply = "reply";
        long highWatermark = 5000;
        MockFProcessor processor = new MockFProcessor(new RuntimeException());
        mockProtocolFactory = new FProtocolFactory(new TJSONProtocol.Factory());
        FNatsServer.Request request = new FNatsServer.Request(data, timestamp, reply, highWatermark,
                mockProtocolFactory, mockProtocolFactory, processor, mockConn);

        request.run();

        byte[] expected = new byte[]{0, 0, 0, 0};
        verify(mockConn).publish(reply, expected);
    }

    @Test
    public void testRequestProcess_noResponse() throws TException, IOException {
        byte[] data = "xxxxhello".getBytes();
        long timestamp = System.currentTimeMillis();
        String reply = "reply";
        long highWatermark = 5000;
        MockFProcessor processor = new MockFProcessor(data, null);
        mockProtocolFactory = new FProtocolFactory(new TJSONProtocol.Factory());
        FNatsServer.Request request = new FNatsServer.Request(data, timestamp, reply, highWatermark,
                mockProtocolFactory, mockProtocolFactory, processor, mockConn);

        request.run();

        verify(mockConn, times(0)).publish(any(String.class), any(byte[].class));
    }

    private class MockFProcessor implements FProcessor {

        private byte[] expectedIn;
        private byte[] expectedOut;
        private RuntimeException runtimeException;

        public MockFProcessor(byte[] expectedIn, byte[] expectedOut) {
            this.expectedIn = expectedIn;
            this.expectedOut = expectedOut;
        }

        public MockFProcessor(RuntimeException runtimeException) {
            this.runtimeException = runtimeException;
        }

        @Override
        public void process(FProtocol in, FProtocol out) throws TException {
            if (runtimeException != null) {
                throw runtimeException;
            }

            assertTrue(in.getTransport() instanceof TMemoryInputTransport);

            if (expectedIn != null) {
                TMemoryInputTransport transport = (TMemoryInputTransport) in.getTransport();
                assertArrayEquals(Arrays.copyOfRange(expectedIn, 4, expectedIn.length), transport.getBuffer());
            }

            if (expectedOut != null) {
                out.writeString(new String(expectedOut));
            }
        }

        @Override
        public void addMiddleware(ServiceMiddleware middleware) {
        }

        @Override
        public Map<String, Map<String, String>> getAnnotations() {
            return null;
        }
    }

}
