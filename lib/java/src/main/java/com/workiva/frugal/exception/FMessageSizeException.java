package com.workiva.frugal.exception;

import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.transport.TTransportException;


/**
 * This exception indicates a message was too large for a transport to handle.
 */
public class FMessageSizeException extends TTransportException {

    public FMessageSizeException() {
        super(FTransport.REQUEST_TOO_LARGE);
    }

    public FMessageSizeException(String message) {
        super(FTransport.REQUEST_TOO_LARGE, message);
    }

    public FMessageSizeException(int type, String message) {
        super(type, message);
    }

    public FMessageSizeException(Throwable cause) {
        super(FTransport.REQUEST_TOO_LARGE, cause);
    }

    public FMessageSizeException(String message, Throwable cause) {
        super(FTransport.REQUEST_TOO_LARGE, message, cause);
    }

}
