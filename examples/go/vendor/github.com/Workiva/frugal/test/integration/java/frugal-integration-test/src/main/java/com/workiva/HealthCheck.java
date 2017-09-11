package com.workiva;

import fi.iki.elonen.NanoHTTPD;
import java.io.IOException;


public class HealthCheck extends NanoHTTPD {

    public HealthCheck(int healthcheckPort) throws IOException {
        super(healthcheckPort);
        start(NanoHTTPD.SOCKET_READ_TIMEOUT, false);
        System.out.println("Starting healthcheck server");
    }

    @Override
    public Response serve(IHTTPSession session) {
        System.out.println("Healthcheck received");
        // All the healthcheck needs is a 200 response, data is irrelevant
        return newFixedLengthResponse("foobar");
    }


}
