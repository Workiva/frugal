# TODO: Remove this with 2.0
from frugal.tornado.transport import FTornadoTransport
from frugal.util.deprecate import deprecated


class TTornadoTransportBase(FTornadoTransport):
    """
    @deprecated Use FTornadoTransport instead
    """

    @deprecated
    def __init__(self, max_message_size=1024 * 1024):
        super(TTornadoTransportBase, self).__init__(
            max_message_size=max_message_size)
