package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.protocol.HeaderUtils;
import com.workiva.frugal.util.Pair;
import com.workiva.frugal.util.ProtocolUtils;

import org.apache.thrift.TConfiguration;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TProtocolException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.After;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.mockito.invocation.InvocationOnMock;

import java.io.UnsupportedEncodingException;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.IntStream;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.fail;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.doAnswer;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FTransport}.
 */
public class FAsyncTransportTest {

    private FAsyncTransportPayloadCapture transport;

    /**
     * Returns a mock message frame.
     */
    public static byte[] mockFrame(FContext context) throws TException, UnsupportedEncodingException {
        byte[] headers = HeaderUtils.encode(context.getRequestHeaders());
        byte[] message = "hello world".getBytes("UTF-8");
        byte[] frame = new byte[headers.length + message.length];
        System.arraycopy(headers, 0, frame, 0, headers.length);
        System.arraycopy(message, 0, frame, headers.length, message.length);
        return frame;
    }


    @Before
    public void setUp() throws Exception {
        transport = new FAsyncTransportPayloadCapture();
    }

    @After
    public void tearDown() throws Exception {
        transport.close();
    }

    /**
     * Ensures request registers context, calls RequestFlusher, returns response, and finally unregisters context.
     */
    @Test
    public void testRequest() throws TException, UnsupportedEncodingException {
        Map<Long, BlockingQueue<byte[]>> mockMap = mock(Map.class);
        transport.queueMap = mockMap;
        transport.open();

        FContext context = new FContext();
        byte[] expectedResponse = FAsyncTransportTest.mockFrame(context);
        doAnswer((InvocationOnMock invocationOnMock) -> {
            BlockingQueue<byte[]> queue = invocationOnMock.getArgument(1);
            queue.put(expectedResponse);
            return null;
        }).when(mockMap).put(eq(FAsyncTransport.getOpId(context)), any());

        byte[] request = "hello world".getBytes();
        TTransport transport = this.transport.request(context, request);
        assertEquals(Integer.MAX_VALUE, transport.getConfiguration().getMaxMessageSize());
        assertArrayEquals(expectedResponse, transport.getBuffer());
        assertArrayEquals(request, this.transport.payloads.get(0));
    }

    /**
     * Ensures request calls flushOp rather than flush.
     */
    @Test
    public void testFlushOp() throws TException, UnsupportedEncodingException {
        FContext context = new FContext();
        byte[] expectedResponse = FAsyncTransportTest.mockFrame(context);
        List<byte[]> payloads = new ArrayList<>();
        FAsyncTransport transport = new FAsyncTransport() {
            @Override
            protected void flush(byte[] payload) throws TTransportException {
                throw new UnsupportedOperationException();
            }

            @Override
            protected void flushOp(long opId, byte[] payload) throws TTransportException {
                payloads.add(payload);
                queueMap.get(opId).add(expectedResponse);
            }
        };
        transport.open();

        byte[] request = "hello world".getBytes();
        assertArrayEquals(expectedResponse, transport.request(context, request).getBuffer());
        assertArrayEquals(request, payloads.get(0));
    }

    /**
     * Ensures request calls flushOp rather than flush.
     */
    @Test
    public void testServiceNotAvailable() throws TException, UnsupportedEncodingException {
        FContext context = new FContext();
        List<byte[]> payloads = new ArrayList<>();
        FAsyncTransport transport = new FAsyncTransport() {
            @Override
            protected void flush(byte[] payload) throws TTransportException {
                throw new UnsupportedOperationException();
            }

            @Override
            protected void flushOp(long opId, byte[] payload) throws TTransportException {
                payloads.add(payload);
                try {
                    handleServiceNotAvailable(opId);
                } catch (TException e) {
                    throw new RuntimeException(e);
                }
            }
        };
        transport.open();

        byte[] request = "hello world".getBytes();
        try {
            transport.request(context, request);
            fail();
        } catch (TTransportException e) {
            assertEquals(TTransportExceptionType.SERVICE_NOT_AVAILABLE, e.getType());
        }
        assertArrayEquals(request, payloads.get(0));
    }

    /**
     * Ensures oneway request calls RequestFlusher and returns null.
     */
    @Test
    public void testOneway() throws TTransportException {
        Map<Long, BlockingQueue<byte[]>> mockMap = mock(Map.class);
        transport.queueMap = mockMap;

        FContext context = new FContext();
        byte[] request = "hello world".getBytes();
        transport.open();
        transport.oneway(context, request);
        transport.close();
        assertArrayEquals(request, transport.payloads.get(0));

        verify(mockMap, times(0)).put(any(), any());
    }

    /**
     * Ensures request timeout throws TTransportException.
     */
    @Test(expected = TTransportException.class)
    public void testRequestTimeout() throws TTransportException {
        FContext context = new FContext();
        context.setTimeout(10);
        transport.open();
        transport.request(context, "hello world".getBytes());
    }

    /**
     * Ensures TTransportException is thrown if poison pill placed in registered queue.
     */
    @Test(expected = TTransportException.class)
    public void testRequestPoisonPill() throws TTransportException {
        Map<Long, BlockingQueue<byte[]>> mockMap = mock(Map.class);
        transport.queueMap = mockMap;
        transport.open();

        FContext context = new FContext();
        doAnswer((InvocationOnMock invocationOnMock) -> {
            BlockingQueue<byte[]> queue = invocationOnMock.getArgument(1);
            queue.put(FAsyncTransport.POISON_PILL);
            return null;
        }).when(mockMap).put(eq(FAsyncTransport.getOpId(context)), any());
        transport.request(context, "hello world".getBytes());
    }

    /**
     * Ensures TTransportException is thrown if a request is already in flight for a given context.
     */
    @Test(expected = TTransportException.class)
    public void testRequestThrowExceptionWithContextInFlight() throws Exception {
        // given
        FContext context = new FContext();

        // when
        transport.queueMap.put(FAsyncTransport.getOpId(context), new ArrayBlockingQueue<>(1));
        transport.open();

        // then (exception)
        transport.request(context, "crap".getBytes());
    }

    @Test
    public void testCloseKillsInFlightRequests() throws Exception {
        // given
        CountDownLatch interruptSignal = new CountDownLatch(1);

        FContext context = new FContext();
        final BlockingQueue<Long> opIds = new ArrayBlockingQueue<>(1);
        FAsyncTransportOpIdQueue tr = new FAsyncTransportOpIdQueue(opIds);
        tr.open();

        ThreadPoolExecutor executorService = (ThreadPoolExecutor) Executors.newCachedThreadPool();
        executorService.execute(() -> {
            try {
                byte[] request = new byte[4];
                ProtocolUtils.writeInt((int) FAsyncTransport.getOpId(context), request, 0);
                tr.request(context, request);
            } catch (TTransportException e) {
                if (e.getType() != TTransportExceptionType.NOT_OPEN) {
                    fail();
                }
            }
            interruptSignal.countDown();
        });

        // Wait for flush to be called
        opIds.take();

        // when
        tr.close();

        // then (success when thread interrupted)
        interruptSignal.await(); // wait for thread interrupt
        assertEquals(tr.queueMap.size(), 0);
    }

    /**
     * Run a producer with multiple consumers.
     * All data requested must have an appropriate response to pass the test.
     * <p>
     * Note:
     * This test may unfairly synchronize consumers by pulling work from the same queue.
     * However, a shared-queue is indicative of real-world use.
     */
    @Test
    public void testTransportIsThreadsafe() throws TException {
        final long poisonPill = Long.MAX_VALUE;

        final ExecutorService pool = Executors.newCachedThreadPool();

        // At test completion, values requested and handled must match
        final AtomicLong requestedSum = new AtomicLong(0);
        final AtomicLong handledSum = new AtomicLong(0);

        final int nRequests = 100_000; // Number of requests to make to the transport
        final int nConsumers = 100; // Number of concurrent responders
        final CyclicBarrier barrier = new CyclicBarrier(nConsumers + 1 + 1); // + 1 producer, + 1 for main thread;
        final BlockingQueue<Long> opIds = new ArrayBlockingQueue<>(nRequests); // Store all operations requested

        FAsyncTransportOpIdQueue tr = new FAsyncTransportOpIdQueue(opIds);
        tr.open();

        class Producer implements Runnable {
            @Override
            public void run() {
                try {
                    barrier.await();

                    IntStream
                            .range(0, nRequests)
                            .forEach(i -> {
                                try {
                                    makeRequest();
                                } catch (InterruptedException e) {
                                    fail();
                                }
                            });

                    // Signal end of queue with poison pill
                    opIds.add(poisonPill);

                    barrier.await();
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }

            private void makeRequest() throws InterruptedException {
                FContext context = new FContext();
                try {
                    byte[] request = new byte[4];
                    ProtocolUtils.writeInt((int) FAsyncTransport.getOpId(context), request, 0);
                    tr.request(context, request);
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

                requestedSum.getAndAdd(FAsyncTransport.getOpId(context));
            }
        }

        class Consumer implements Runnable {

            @Override
            public void run() {
                try {
                    barrier.await();

                    while (true) {
                        long opId = opIds.take();
                        if (opId == poisonPill) {
                            opIds.put(opId); // notify other threads to quit
                            barrier.await(); // release barrier
                            return;
                        }

                        FContext context = new FContext();
                        context.addRequestHeader(FContext.OPID_HEADER, Long.toString(opId));

                        tr.handleResponse(mockFrame(context));

                        handledSum.getAndAdd(FAsyncTransport.getOpId(context));
                    }

                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }
        }

        try {
            pool.execute(new Producer());
            IntStream.range(0, nConsumers).forEach(i -> pool.execute(new Consumer()));

            barrier.await(); // wait for all threads to be ready
            barrier.await(); // wait for all threads to finish

            assertEquals(requestedSum.get(), handledSum.get());
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
        // close the transport
        tr.close();
    }

    /**
     * Ensures handleResponse throws TProtocolException if opid format is bad.
     */
    @Test(expected = TProtocolException.class)
    public void testHandleResponseBadOpId() throws TException, UnsupportedEncodingException {
        FContext ctx = new FContext();
        ctx.addRequestHeader(FContext.OPID_HEADER, "foo");
        transport.handleResponse(mockFrame(ctx));
    }

    /**
     * Ensures handleResponse drops unregistered responses.
     */
    @Test
    public void testHandleResponseDropsUnregisteredOpId() throws TException, UnsupportedEncodingException {
        // given
        transport.queueMap = spy(transport.queueMap);

        // when
        transport.handleResponse(mockFrame(new FContext()));

        // then
        verify(transport.queueMap, times(1)).get(any());
    }

    class FAsyncTransportOpIdQueue extends FAsyncTransport {
        final BlockingQueue<Long> opIds;

        FAsyncTransportOpIdQueue(BlockingQueue<Long> opIds) {
            this.opIds = opIds;
        }

        @Override
        protected void flush(byte[] payload) throws TTransportException {
            try {
                opIds.put((long) ProtocolUtils.readInt(payload, 0));
            } catch (InterruptedException e) {
                throw new TTransportException(e);
            }

        }
    }

    class FAsyncTransportPayloadCapture extends FAsyncTransport {

        ArrayList<byte[]> payloads = new ArrayList<>(1);

        @Override
        protected void flush(byte[] payload) throws TTransportException {
            this.payloads.add(payload);
        }
    }
}
