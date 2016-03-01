from threading import Lock

import base.BaseFoo

from frugal.registry import FClientRegistry
from . import Foo


class Iface(base.BaseFoo.Iface):

    def one_way(self, id, req):
        pass


class Client(base.BaseFoo.Client, Iface):

    def __init__(self, transport, protocol_factory):
        """Initialize a Client with a transport and protocol factory creating a
        new FClientRegistry

            Args:
                transport: FTransport
                protocol_factory: FProtocolFactory
        """
        base.BaseFoo.Client.__init__(self, transport, protocol_factory)
        self._transport = transport
        self._transport.set_registry(FClientRegistry())
        self._protocol_factory = protocol_factory
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def one_way(self, ctx, id, req):
        oprot = self._oprot
        with self._write_lock:
            oprot.write_request_headers(ctx)
            args = Foo.one_way_args()
            args.id = id
            args.req = req
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()
