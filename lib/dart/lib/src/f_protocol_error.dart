part of frugal;

class FProtocolError extends TProtocolError {
  FProtocolError([int type = TProtocolErrorType.UNKNOWN, String message = ""])
      : super(type, message);
}
