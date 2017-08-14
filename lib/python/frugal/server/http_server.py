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

from base64 import b64encode
from base64 import b64decode
import logging
from struct import pack

from six.moves import BaseHTTPServer
from thrift.Thrift import TApplicationException
from thrift.transport import TTransport

from frugal.server import FServer

logger = logging.getLogger(__name__)


class FHttpServer(FServer):
    """
    Simple FServer implementation using HTTP. This is merely a reference
    implementation and is not production-ready.
    """

    def __init__(self, processor, address, proto_factory):
        """
        Initialize an FHttpServer.

        Args:
            processor: FProcessor used to process requests.
            address: tuple of host name and port.
            proto_factory: FProtocolFactory used to read requests and write
                           responses.
        """

        class RequestHandler(BaseHTTPServer.BaseHTTPRequestHandler):
            """
            Request handler implements a frugal HTTP handler.
            """

            def do_POST(self):
                """
                do_POST implements POST method on a frugal server.
                """
                length = self.headers.getheader('Content-Length')
                payload = b64decode(self.rfile.read(int(length)))
                response_limit = int(self.headers.getheader(
                    'x-frugal-payload-limit', 0))

                if len(payload) <= 4:
                    logging.exception(
                        'Invalid request frame length %s', len(payload))
                    self.send_response(400)
                    self.end_headers()
                    return

                itrans = TTransport.TMemoryBuffer(payload[4:])
                otrans = TTransport.TMemoryBuffer()
                iprot = proto_factory.get_protocol(itrans)
                oprot = proto_factory.get_protocol(otrans)

                try:
                    processor.process(iprot, oprot)
                except TApplicationException:
                    # Continue so the exception is sent to the client
                    pass
                except Exception:
                    self.send_response(400)
                    self.end_headers()
                    return

                # Encode response.
                response = otrans.getvalue()
                if len(response) > response_limit > 0:
                    self.send_response(413)
                    self.end_headers()
                    return

                frame_length = pack('!I', len(response))
                frame = b64encode(frame_length + response)

                self.send_response(200)
                self.send_header('Content-Type', 'application/x-frugal')
                self.send_header('Content-Length', len(frame))
                self.send_header('Content-Transfer-Encoding', 'base64')
                self.end_headers()

                self.wfile.write(frame)

        self._httpd = BaseHTTPServer.HTTPServer(address, RequestHandler)

    def serve(self):
        self._httpd.serve_forever()

    def stop(self):
        self._httpd.socket.close()
