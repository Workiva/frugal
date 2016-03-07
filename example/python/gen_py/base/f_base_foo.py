from Queue import Queue

from frugal.registry import FClientRegistry


class Iface(object):

    def base_ping(context):
        pass


class Client(Iface):

    def __init__(self, transport, protocol_factory):
        """Initialize a Frugal Client

        Args:
            transport: FTransport
            protocl_factory: FProtocolFactory for creating FProtocols
        """

        self._transport = transport
        self._transport.set_registry(FClientRegistry())
        self._protocol_factory = protocol_factory
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._oport = self._protocol_factory.get_protocol(self._transport)

    def base_ping(self, ctx):
        """ base ping

        Args:
            ctx: FContext
        """
        oprot = self._oport
        result = Queue(1)
        self._transport.register(ctx, self._recv_base_ping_handler(ctx, result))

        with self._write_lock:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin()
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _recv_base_ping_handler(self, context, result):
        pass


