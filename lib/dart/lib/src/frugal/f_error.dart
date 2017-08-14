/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

part of frugal.src.frugal;

/// Contains [TApplicationError] types used in frugal instantiated
/// [TApplicationError]s.
class FrugalTApplicationErrorType extends TApplicationErrorType {
  /// Inherited from thrift.
  static const int UNKNOWN = TApplicationErrorType.UNKNOWN;

  /// Inherited from thrift.
  static const int UNKNOWN_METHOD = TApplicationErrorType.UNKNOWN_METHOD;

  /// Inherited from thrift.
  static const int INVALID_MESSAGE_TYPE =
      TApplicationErrorType.INVALID_MESSAGE_TYPE;

  /// Inherited from thrift.
  static const int WRONG_METHOD_NAME = TApplicationErrorType.WRONG_METHOD_NAME;

  /// Inherited from thrift.
  static const int BAD_SEQUENCE_ID = TApplicationErrorType.BAD_SEQUENCE_ID;

  /// Inherited from thrift.
  static const int MISSING_RESULT = TApplicationErrorType.MISSING_RESULT;

  /// Inherited from thrift.
  static const int INTERNAL_ERROR = TApplicationErrorType.INTERNAL_ERROR;

  /// Inherited from thrift.
  static const int PROTOCOL_ERROR = TApplicationErrorType.PROTOCOL_ERROR;

  /// Inherited from thrift.
  static const int INVALID_TRANSFORM = TApplicationErrorType.INVALID_TRANSFORM;

  /// Inherited from thrift.
  static const int INVALID_PROTOCOL = TApplicationErrorType.INVALID_PROTOCOL;

  /// Inherited from thrift.
  static const int UNSUPPORTED_CLIENT_TYPE =
      TApplicationErrorType.UNSUPPORTED_CLIENT_TYPE;

  /// Indicates the response was too large for the transport.
  static const int RESPONSE_TOO_LARGE = 100;
}

/// Contains [TTransportError] types used in frugal instantiated
/// [TTransportError]s.
class FrugalTTransportErrorType extends TTransportErrorType {
  /// Inherited from thrift.
  static const int UNKNOWN = TTransportErrorType.UNKNOWN;

  /// Inherited from thrift.
  static const int NOT_OPEN = TTransportErrorType.NOT_OPEN;

  /// Inherited from thrift.
  static const int ALREADY_OPEN = TTransportErrorType.ALREADY_OPEN;

  /// Inherited from thrift.
  static const int TIMED_OUT = TTransportErrorType.TIMED_OUT;

  /// Inherited from thrift.
  static const int END_OF_FILE = TTransportErrorType.END_OF_FILE;

  /// Indicates the request was too large for the transport.
  static const int REQUEST_TOO_LARGE = 100;

  /// Indicates the response was too large for the transport.
  static const int RESPONSE_TOO_LARGE = 101;
}
