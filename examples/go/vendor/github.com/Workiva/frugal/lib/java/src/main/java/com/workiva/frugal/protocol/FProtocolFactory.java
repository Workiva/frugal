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

import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransport;

/**
 * FProtocolFactory creates new FProtocol instances. It takes a TProtocolFactory
 * and a TTransport and returns an FProtocol which wraps a TProtocol produced by
 * the TProtocolFactory. The TProtocol itself wraps the provided TTransport. This
 * makes it easy to produce an FProtocol which uses any existing Thrift transports
 * and protocols in a composable manner.
 */
public class FProtocolFactory {

    private TProtocolFactory tProtocolFactory;

    public FProtocolFactory(TProtocolFactory tProtocolFactory) {
        this.tProtocolFactory = tProtocolFactory;
    }

    public FProtocol getProtocol(TTransport transport) {
        return new FProtocol(tProtocolFactory.getProtocol(transport));
    }

}
