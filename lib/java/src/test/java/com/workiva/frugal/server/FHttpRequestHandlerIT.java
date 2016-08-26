package com.workiva.frugal.server;

import com.workiva.frugal.IntegrationTest;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FHttpTransport;
import org.apache.http.ConnectionClosedException;
import org.apache.http.config.SocketConfig;
import org.apache.http.impl.bootstrap.HttpServer;
import org.apache.http.impl.bootstrap.ServerBootstrap;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.experimental.categories.Category;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;
import v1.music.Album;
import v1.music.FStore;
import v1.music.PerfRightsOrg;
import v1.music.PurchasingError;
import v1.music.Track;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.fail;

/**
 * Integration tests for {@link FHttpRequestHandler}.
 */
@Category(IntegrationTest.class)
@RunWith(JUnit4.class)
public class FHttpRequestHandlerIT {

    public static HttpServer server;
    public static Thread serverThread;

    @BeforeClass
    public static void startHttpServer() throws IOException, InterruptedException {
        SocketConfig socketConfig = SocketConfig.custom()
                .setSoTimeout(15000)
                .setTcpNoDelay(true)
                .build();

        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());
        FStore.Processor processor = new FStore.Processor(new FStoreHandler());
        server = ServerBootstrap.bootstrap()
                .setListenerPort(8080)
                .setServerInfo("Test/1.1")
                .setSocketConfig(socketConfig)
                .setExceptionLogger(ex -> {
                    if (ex instanceof SocketTimeoutException) {
                        System.err.println("Connection timed out");
                    } else if (ex instanceof ConnectionClosedException) {
                        System.err.println("Error: " + ex.getMessage());
                    } else {
                        ex.printStackTrace();
                    }
                })
                .registerHandler("*", FHttpRequestHandler.of(processor, protocolFactory, protocolFactory))
                .create();

        // Start the server in a new thread
        serverThread = new Thread(() -> {
            try {
                server.start();
            } catch (Exception e) {
                fail(e.getMessage());
            }
        });
        serverThread.start();

        // Halt thread if JVM stops
        Runtime.getRuntime().
                addShutdownHook(new Thread() {
                    @Override
                    public void run() {
                        server.shutdown(5, TimeUnit.SECONDS);
                    }
                });
    }

    private static class FStoreHandler implements FStore.Iface {

        @Override
        public Album buyAlbum(FContext ctx, String asin, String acct) throws TException, PurchasingError {
            Album album = new Album();
            album.setASIN(asin);
            album.setDuration(1200);
            album.addToTracks(
                    new Track(
                            "Comme des enfants",
                            "Coeur de pirate",
                            "Grosse Boîte",
                            "Béatrice Martin",
                            169,
                            PerfRightsOrg.ASCAP));
            return album;
        }

        @Override
        public boolean enterAlbumGiveaway(FContext ctx, String email, String name) throws TException {
            return true;
        }
    }


    @AfterClass
    public static void stopServer() throws InterruptedException {
        server.awaitTermination(1, TimeUnit.SECONDS);
        serverThread.join();
    }

    @Test
    public void testParallelClients() throws TException, InterruptedException, IOException {
        CloseableHttpClient httpClient = HttpClients.createDefault();
        FHttpTransport transport = new FHttpTransport.Builder(httpClient, "http://localhost:8080").build();
        FProtocolFactory protocolFactory = new FProtocolFactory(new TBinaryProtocol.Factory());
        FStore.Client storeClient = new FStore.Client(transport, protocolFactory);

        Album album = storeClient.buyAlbum(new FContext("corr-id-1"), "ASIN-1290AIUBOA89", "ACCOUNT-12345");
        assertEquals(album.getASIN(), "ASIN-1290AIUBOA89");
        assertEquals(album.getDuration(), 1200, 0.01);
        Track track = album.getTracks().get(0);
        assertEquals(track.getTitle(), "Comme des enfants");
        assertEquals(track.getArtist(), "Coeur de pirate");
        assertEquals(track.getPublisher(), "Grosse Boîte");
        assertEquals(track.getComposer(), "Béatrice Martin");
        assertEquals(track.getDuration(), 169, 0.01);
        assertEquals(track.getPro(), PerfRightsOrg.ASCAP);

        storeClient.enterAlbumGiveaway(new FContext("corr-id-2"), "kevin@workiva.com", "Kevin");

        httpClient.close();
        transport.close();
    }

}
