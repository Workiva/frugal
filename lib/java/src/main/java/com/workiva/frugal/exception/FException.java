package com.workiva.frugal.exception;

import org.apache.thrift.TException;

public class FException extends TException {

    public FException() {
        super();
    }

    public FException(String message) {
        super(message);
    }

    public FException(Throwable cause) {
        super(cause);
    }

    public FException(String message, Throwable cause) {
        super(message, cause);
    }
}
