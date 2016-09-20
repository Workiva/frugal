package com.workiva.frugal.integration;

import com.workiva.frugal.protocol.FContext;
import v1.music.Album;
import v1.music.FStore;
import v1.music.PerfRightsOrg;
import v1.music.PurchasingError;
import v1.music.Track;
import org.apache.thrift.TException;


/**
 * A handler handles all incoming requests to the server.
 * The handler must satisfy the interface the server exposes.
 */
public class FStoreHandler implements FStore.Iface {

    /**
     * Return an album; always buy the same one.
     */
    @Override
    public Album buyAlbum(FContext ctx, String ASIN, String acct) throws TException, PurchasingError {
        Album album = new Album();
        album.setASIN("12345");
        album.setDuration(2000);
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