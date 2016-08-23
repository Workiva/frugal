import base64
import struct

from aiohttp import web
from thrift.transport.TTransport import TMemoryBuffer

from frugal.aio.processor import FProcessor
from frugal.protocol import FProtocolFactory


def new_http_handler(processor: FProcessor, protocol_factory: FProtocolFactory):
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
        out_transport = TMemoryBuffer()
        oprot = protocol_factory.get_protocol(out_transport)
        try:
            await processor.process(iprot, oprot)
        except:
            # TODO make this less broad in the future
            return web.Response(status=400)

        # write back response
        output_data = out_transport.getvalue()
        if len(output_data) > response_limit > 0:
            return web.Response(status=413)

        output_data_len = struct.pack('!I', len(output_data))
        output_payload = base64.b64encode(output_data_len + output_data)

        return web.Response(body=output_payload, headers=headers)

    return handler
