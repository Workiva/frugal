package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import org.apache.thrift.TException;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.io.UnsupportedEncodingException;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.IntStream;

import static junit.framework.TestCase.fail;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.spy;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FClientRegistry}.
 */
@RunWith(JUnit4.class)
public class FClientRegistryTest {

    /**
     * Returns a mock message frame.
     */
    private byte[] mockFrame(FContext context) throws TException, UnsupportedEncodingException {
        byte[] headers = HeaderUtils.encode(context.getRequestHeaders());
        byte[] message = "hello world".getBytes("UTF-8");
        byte[] frame = new byte[headers.length + message.length];
        System.arraycopy(headers, 0, frame, 0, headers.length);
        System.arraycopy(message, 0, frame, headers.length, message.length);
        return frame;
    }

    @Test
    public void testRegisterAddsToCallbackList() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();
        FContext context = new FContext();

        // when
        registry.register(context, transport -> { });

        // then
        assertEquals(1, registry.handlers.size());
    }

    @Test(expected = FException.class)
    public void testRegisterThrowsExceptionForMultipleOpIds() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();
        FContext context = new FContext();
        FAsyncCallback callback = transport -> { };

        // when
        registry.register(context, callback);

        // then (exception)
        registry.register(context, callback);
    }

    @Test
    public void testRegisterAssignsOpId() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();
        FContext context = new FContext();

        // when
        registry.register(context, transport -> {});

        // then
        assertNotEquals(context.getOpId(), 0);
    }

    @Test
    public void testRegisterAssignsIncreasingOpId() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();

        // when
        FContext context1 = new FContext();
        registry.register(context1, transport -> {});
        FContext context2 = new FContext();
        registry.register(context2, transport -> {});

        // then
        assertNotEquals(context1.getOpId(), context2.getOpId());
    }

    @Test
    public void testUnregisterRemovesFromHandlers() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();
        FContext context = new FContext();
        registry.register(context, transport -> {});

        // when
        registry.unregister(context);

        // then
        assertEquals(0, registry.handlers.size());
    }

    @Test
    public void testUnregisterMissingContextSucceeds() throws Exception {
        // given
        FClientRegistry registry = new FClientRegistry();

        // when
        registry.unregister(new FContext());
        registry.unregister(null);

        // then (succeeds)
    }

    @Test
    public void testExecuteDropsUnregisteredOpId() throws TException, UnsupportedEncodingException {
        // given
        FClientRegistry registry = new FClientRegistry();
        registry.handlers = spy(registry.handlers);

        FContext context = new FContext();
        byte[] frame = mockFrame(context);

        // when
        registry.execute(frame);

        // then
        verify(registry.handlers, times(1)).get(any());
    }

    @Test
    public void testCloseInterruptsRunningThreads() throws Exception {
        // given
        CountDownLatch readySignal = new CountDownLatch(1);
        CountDownLatch interruptSignal = new CountDownLatch(1);

        FClientRegistry registry = new FClientRegistry();
        FContext context = new FContext();
        ThreadPoolExecutor executorService = (ThreadPoolExecutor) Executors.newCachedThreadPool();
        executorService.execute(() -> {
            try {
                registry.register(context, transport -> { }); // no-op callback
            } catch (TException e) {
                fail();
            }
            readySignal.countDown();

            // spin wait for interrupt signal
            while (!Thread.currentThread().isInterrupted()) { }

            interruptSignal.countDown();
        });

        readySignal.await(); // wait for thread ready

        // when
        registry.close();

        // then (success when thread interrupted)
        interruptSignal.await(); // wait for thread interrupt
        assertEquals(registry.handlers.size(), 0);
    }

    /**
     * Run a producer with multiple consumers.
     * All data put in to the registry must correctly be consumed to pass the test.
     * <p>
     * Note:
     * This test may unfairly synchronize consumers by pulling work from the same queue.
     * However, a shared-queue is indicative of real-world use.
     */
    @Test
    public void testRegistryIsThreadsafe() throws TException {
        final long poisonPill = Long.MAX_VALUE;

        final ExecutorService pool = Executors.newCachedThreadPool();

        // At test completion, values registered, unregistered, and executed must match
        final AtomicLong registerSum = new AtomicLong(0);
        final AtomicLong unregisterSum = new AtomicLong(0);
        final AtomicLong executeSum = new AtomicLong(0);

        final int nRegistrations = 100_000; // Number of registrations to make to the registry
        final int nConsumers = 100; // Number of concurrent consumers
        final CyclicBarrier barrier = new CyclicBarrier(nConsumers + 1 + 1); // + 1 producer, + 1 for main thread;
        final BlockingQueue<Long> opIds = new ArrayBlockingQueue<>(nRegistrations); // Store all operations registered
        final FClientRegistry registry = new FClientRegistry();

        class Producer implements Runnable {
            @Override
            public void run() {
                try {
                    barrier.await();

                    IntStream
                            .range(0, nRegistrations)
                            .forEach(i -> putToRegistry());

                    // Signal end of queue with poison pill
                    opIds.add(poisonPill);

                    barrier.await();
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }

            private void putToRegistry() {
                FContext context = new FContext();

                try {
                    registry.register(context, transport -> executeSum.getAndAdd(context.getOpId()));
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

                opIds.add(context.getOpId());
                registerSum.getAndAdd(context.getOpId());
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
                        context.setOpId(opId);

                        byte[] frame = mockFrame(context);
                        registry.execute(frame);
                        registry.unregister(context);

                        unregisterSum.getAndAdd(context.getOpId());
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

            assertEquals(registerSum.get(), unregisterSum.get());
            assertEquals(registerSum.get(), executeSum.get());
            assertEquals(registry.handlers.size(), 0);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }

    }

}
