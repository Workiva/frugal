package com.workiva.frugal.exception;

import org.apache.thrift.TApplicationException;

/**
 * Contains TApplicationException types used in frugal instantiated TApplicationExceptions.
 */
public class FrugalTApplicationExceptionType {

    // Thrift-inherited types.

    public static final int UNKNOWN = TApplicationException.UNKNOWN;
    public static final int UNKNOWN_METHOD = TApplicationException.UNKNOWN_METHOD;
    public static final int INVALID_MESSAGE_TYPE = TApplicationException.INVALID_MESSAGE_TYPE;
    public static final int WRONG_METHOD_NAME = TApplicationException.WRONG_METHOD_NAME;
    public static final int BAD_SEQUENCE_ID = TApplicationException.BAD_SEQUENCE_ID;
    public static final int MISSING_RESULT = TApplicationException.MISSING_RESULT;
    public static final int INTERNAL_ERROR = TApplicationException.INTERNAL_ERROR;
    public static final int PROTOCOL_ERROR = TApplicationException.PROTOCOL_ERROR;
    public static final int INVALID_TRANSFORM = TApplicationException.INVALID_TRANSFORM;
    public static final int INVALID_PROTOCOL = TApplicationException.INVALID_PROTOCOL;
    public static final int UNSUPPORTED_CLIENT_TYPE = TApplicationException.UNSUPPORTED_CLIENT_TYPE;


    // Frugal-specific types.

    /**
     * Indicates the response was too large for the transport.
     */
    public static final int RESPONSE_TOO_LARGE = 100;
}
