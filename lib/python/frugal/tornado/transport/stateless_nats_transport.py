# TODO: Remove with 2.0
from frugal.tornado.transport import FNatsTransport
from frugal.util.deprecate import deprecated


class TStatelessNatsTransport(FNatsTransport):
    """TStatelessNatsTransport is an extension of thrift.TTransportBase.
    This is a "stateless" transport in the sense that there is no
    connection with a server. A request is simply published to a subject
    and responses are received on another subject. This assumes
    requests/responses fit within a single NATS message.

    @deprecated - Use FNatsTransport instead
    """

    @deprecated
    def __init__(self, nats_client, subject, inbox=""):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        super(TStatelessNatsTransport, self).__init__(
            nats_client=nats_client,
            subject=subject,
            inbox=inbox,
            is_ttransport=True,
        )
