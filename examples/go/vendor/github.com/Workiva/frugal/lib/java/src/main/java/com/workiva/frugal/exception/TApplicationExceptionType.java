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

package com.workiva.frugal.exception;

import org.apache.thrift.TApplicationException;

/**
 * Contains TApplicationException types used in frugal instantiated TApplicationExceptions.
 */
public class TApplicationExceptionType {

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
