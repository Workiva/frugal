package com.workiva.frugal.benchmarks;

import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import io.nats.client.Nats;
import org.openjdk.jmh.annotations.Benchmark;
import org.openjdk.jmh.annotations.Scope;
import org.openjdk.jmh.annotations.Setup;
import org.openjdk.jmh.annotations.State;
import org.openjdk.jmh.annotations.TearDown;
import org.openjdk.jmh.runner.Runner;
import org.openjdk.jmh.runner.RunnerException;
import org.openjdk.jmh.runner.options.Options;
import org.openjdk.jmh.runner.options.OptionsBuilder;

import java.io.IOException;

/**
 * Benchmarks for JNATS.
 */
@State(Scope.Thread)
public class NatsBenchmark {

    Connection nc;

    @Setup
    public void setup() throws IOException {
        ConnectionFactory cf = new ConnectionFactory(Nats.DEFAULT_URL);
        try {
            nc = cf.createConnection();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    @TearDown
    public void teardown() {
        nc.close();
    }

    @Benchmark
    public void testPublisher() {
        try {
            nc.publish("topic", "Hello World".getBytes());
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public static void main(String[] args) throws RunnerException {
        Options opt = new OptionsBuilder()
                .include(NatsBenchmark.class.getSimpleName())
                .warmupIterations(5)
                .measurementIterations(5)
                .forks(1)
                .build();

        new Runner(opt).run();
    }
}
