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

import org.apache.thrift.transport.TTransportException;

/**
 * Contains TTransportException types used in frugal instantiated TTransportExceptions.
 */
public class TTransportExceptionType {

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
