package com.workiva.frugal.protocol;

/**
 * Headers used within the Frugal protocol.
 */
public final class Headers {

    public static final String ACCEPT_HEADER = "Accept";
    public static final String CONTENT_TRANSFER_ENCODING_HEADER = "Content-Transfer-Encoding";

    public static final String APPLICATION_X_FRUGAL_HEADER = "Application/X-Frugal";
    public static final String X_FRUGAL_PAYLOAD_LIMIT_HEADER = "X-Frugal-Payload-Limit";
    public static final String CONTENT_TRANSFER_ENCODING = "Base64";
    public static final String CONTENT_TYPE = "UTF-8";

    private Headers() {}

}
