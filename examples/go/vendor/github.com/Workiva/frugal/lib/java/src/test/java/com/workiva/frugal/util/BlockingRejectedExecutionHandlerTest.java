package com.workiva.frugal.util;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;


import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Future;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

import static org.junit.Assert.fail;

/**
 * Tests for {@link BlockingRejectedExecutionHandler}.
 */
@RunWith(JUnit4.class)
public class BlockingRejectedExecutionHandlerTest {

    /**
     * Ensure that with BlockingRejectedExecutionHandler, a runnable submitted to a full queue blocks until the queue
     * has space, and the task is eventually executed.
     */
    @Test
    public void testRejectedExecution() throws InterruptedException {
        BlockingQueue<Runnable> workQueue = new SynchronousQueue<>();
        ExecutorService executor =
                new ThreadPoolExecutor(1, 1, 1, TimeUnit.MILLISECONDS, workQueue,
                        new BlockingRejectedExecutionHandler());

        final CountDownLatch latch = new CountDownLatch(1);
        executor.submit(new Runnable() {
            public void run() {
                try {
                    latch.await();
                    Thread.sleep(50); // Ensure we have a chance to submit the second task with a full queue
                } catch (InterruptedException e) {
                    fail(e.getMessage());
                }
            }
        });

        // Allow first runnable to proceed eventually.
        new Thread(new Runnable() {
            public void run() {
                latch.countDown();
            }
        }).start();

        Future<?> task = executor.submit(new Runnable() {
            public void run() {
                // Don't really care about what goes on in here...
            }
        });

        // Ensure second runnable executes.
        try {
            task.get(1, TimeUnit.SECONDS);
        } catch (ExecutionException e) {
            fail(e.getMessage());
        } catch (TimeoutException e) {
            fail("Second runnable not executed");
        }
    }

}
