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

package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

/**
 * FAsyncCallback is an internal callback which is constructed by generated code
 * and invoked by an FRegistry when a RPC response is received. In other words,
 * it's used to complete RPCs. The operation ID on FContext is used to look up the
 * appropriate callback. FAsyncCallback is passed an in-memory TTransport which
 * wraps the complete message. The callback returns an error or throws an
 * exception if an unrecoverable error occurs and the transport needs to be
 * shutdown.
 */
public interface FAsyncCallback {
    void onMessage(TTransport transport) throws TException;
}
