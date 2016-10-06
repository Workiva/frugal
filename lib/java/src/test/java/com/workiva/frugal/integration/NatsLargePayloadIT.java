package com.workiva.frugal.integration;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.DefaultAWSCredentialsProviderChain;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.AmazonS3Client;
import com.workiva.frugal.IntegrationTest;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.server.FNatsServer;
import com.workiva.frugal.server.FServer;
import com.workiva.frugal.transport.FNatsTransport;
import com.workiva.frugal.transport.FTransport;
import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.transport.TTransportException;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.experimental.categories.Category;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import v1.music.Album;
import v1.music.FStore;

import java.io.IOException;
import java.util.concurrent.TimeoutException;

import static org.junit.Assert.assertEquals;

/**
 * Integration tests for full compiler workflow.
 */
@Category(IntegrationTest.class)
@RunWith(JUnit4.class)
public class NatsLargePayloadIT {

    public static final String SERVICE_SUBJECT = "music-service";
    public static final String S3_BUCKET = "messaging-large-payload";
    private FServer server;
    private Thread serverThread;

    class MusicServer implements Runnable {

        @Override
        public void run() {
            try {
                AWSCredentials credentials = new DefaultAWSCredentialsProviderChain().getCredentials();
                AmazonS3 s3 = new AmazonS3Client(credentials);

                FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

                ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
                Connection conn = cf.createConnection();

                FStore.Processor processor = new FStore.Processor(new FStoreHandler());
                server = new FNatsServer.Builder(conn, processor, protocolFactory, SERVICE_SUBJECT)
                        .withLargePayloadEnabled(s3, S3_BUCKET).build();
                server.serve();
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }

    @Before
    public void setupStoreProcessor() throws IOException, TimeoutException, TTransportException, InterruptedException {
        serverThread = new Thread(new MusicServer());
        serverThread.start();
        Thread.sleep(1000); // wait for server to start
    }

    @After
    public void tearDownStoreProcessor() throws IOException, TimeoutException, TException, InterruptedException {
        assert server != null;
        server.stop();
        assert serverThread != null;
        serverThread.join();
    }

    @Test
    public void testBuyAlbum() throws TException, IOException, TimeoutException {
        // given
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());

        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        Connection conn = cf.createConnection();

        AWSCredentials credentials = new DefaultAWSCredentialsProviderChain().getCredentials();
        AmazonS3 s3 = new AmazonS3Client(credentials);

        FTransport transport = FNatsTransport
                .of(conn, SERVICE_SUBJECT)
                .withLargePayloadEnabled(s3, S3_BUCKET);
        transport.open();

        FStore.Client storeClient = new FStore.Client(transport, protocolFactory);

        // when
        FContext ctx = new FContext();
        ctx.setTimeout(60 * 1000);
        Album album = storeClient.buyAlbum(ctx, "ASIN-1290AIUBOA89", "ACCOUNT-12345");

        // then
        assertEquals(album.ASIN, "12345");
        assertEquals(album.duration, 2000, 0.1);
        assertEquals(album.tracks.size(), 1);

        transport.close();
        conn.close();
    }

}