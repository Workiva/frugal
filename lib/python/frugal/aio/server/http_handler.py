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

import base64

from aiohttp import web
from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TMemoryBuffer

from frugal.aio.processor import FProcessor
from frugal.protocol import FProtocolFactory
from frugal.transport import TMemoryOutputBuffer


def new_http_handler(processor: FProcessor,
                     protocol_factory: FProtocolFactory):
    """
    Returns a function that can be used as a request handler in an aiohttp
    web server.

    Args:
        processor: The processor to use to handle requests.
        protocol_factory: A protocol factory to serialize/deserialize
                          frugal requests.
    """
    async def handler(request: web.Request):
        headers = {
            'content-type': 'application/x-frugal',
            'content-transfer-encoding': 'base64',
        }

        # check for response size limit
        response_limit = request.headers.get('x-frugal-payload-limit') or 0
        if response_limit:
            response_limit = int(response_limit)

        # decode payload and process
        payload = base64.b64decode(await request.content.read())
        iprot = protocol_factory.get_protocol(TMemoryBuffer(payload[4:]))
        # TODO could be better with this limit
        otrans = TMemoryOutputBuffer(0)
        oprot = protocol_factory.get_protocol(otrans)
        try:
            await processor.process(iprot, oprot)
        except TApplicationException:
            # Continue so the exception is sent to the client
            pass
        except Exception:
            return web.Response(status=400)

        # write back response
        output_data = otrans.getvalue()
        if len(output_data) > response_limit > 0:
            return web.Response(status=413)

        output_payload = base64.b64encode(output_data)

        return web.Response(body=output_payload, headers=headers)

    return handler
