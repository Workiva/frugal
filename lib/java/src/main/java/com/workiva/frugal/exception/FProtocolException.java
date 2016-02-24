package com.workiva.frugal.exception;

import org.apache.thrift.protocol.TProtocolException;

public class FProtocolException extends TProtocolException {

    public FProtocolException() {
        super();
    }

    public FProtocolException(int type) {
        super(type);
    }

    public FProtocolException(int type, String message) {
        super(type, message);
    }

    public FProtocolException(int type, Throwable cause) {
        super(type, cause);
    }

    public FProtocolException(Throwable cause) {
        super(cause);
    }

    public FProtocolException(String message) {
        super(message);
    }

    public FProtocolException(String message, Throwable cause) {
        super(message, cause);
    }

    public FProtocolException(int type, String message, Throwable cause) {
        super(message, cause);
    }
}
