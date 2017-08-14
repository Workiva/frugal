# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from threading import Lock

from thrift.transport.TTransport import TTransportException
from thrift.transport.TTransport import TFramedTransport

from . import FServer


class FSimpleServer(FServer):
    """
    FSimpleServer is a very basic implementation of an FServer.
    """

    def __init__(self, processor_factory, server_transport, protocol_factory):
        self._processor_factory = processor_factory
        self._server_transport = server_transport
        self._protocol_factory = protocol_factory
        self._stopped = False
        self._stopped_mu = Lock()

    def serve(self):
        self._server_transport.listen()
        while True:
            with self._stopped_mu:
                if self._stopped:
                    return

            client = self._server_transport.accept()
            framed = TFramedTransport(client)
            iprot = self._protocol_factory.get_protocol(framed)
            oprot = self._protocol_factory.get_protocol(framed)
            processor = self._processor_factory.get_processor(framed)

            try:
                while True:
                    with self._stopped_mu:
                        if self._stopped:
                            break

                    processor.process(iprot, oprot)
            except TTransportException:
                continue
            except Exception as e:
                print(e)
                break

            framed.close()

    def stop(self):
        self._stopped = True
