package com.workiva.frugal.exception;

import org.apache.thrift.transport.TTransportException;

/**
 * Contains TTransportException types used in frugal instantiated TTransportExceptions.
 */
public class FrugalTTransportExceptionType {

    // Thrift-inherited types.

    public static final int UNKNOWN = TTransportException.UNKNOWN;
    public static final int NOT_OPEN = TTransportException.NOT_OPEN;
    public static final int ALREADY_OPEN = TTransportException.ALREADY_OPEN;
    public static final int TIMED_OUT = TTransportException.TIMED_OUT;
    public static final int END_OF_FILE = TTransportException.END_OF_FILE;

    // Frugal-specific types.

    /**
     * TTransportException code which indicates the request was too large for the transport.
     */
    public static final int REQUEST_TOO_LARGE = 100;

    /**
     * TTransportException code which indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 101;

}
