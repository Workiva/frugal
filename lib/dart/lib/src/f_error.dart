part of frugal;

class FError extends TError {
  FError() : super(0, "");

  FError.withMessage(String message) : super(0, message);

  FError.withType(int type, String message) : super(type, message);
}

/**
 * This exception indicates a message was too large for a transport to handle.
 */
class FMessageSizeError extends TTransportError {
  FMessageSizeError(int type, String message) : super(type, message);

  FMessageSizeError.request()
      : super(FTransport.REQUEST_TOO_LARGE,
            "request was too large for the transport");

  FMessageSizeError.response()
      : super(FTransport.RESPONSE_TOO_LARGE,
            "response was too large for the transport");
}

class FTimeoutException extends FError {}
