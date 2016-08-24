from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportException
from tornado import gen, locks

from frugal.exceptions import FExecuteCallbackNotSet
from frugal.tornado.transport import FTornadoTransport


class FNatsTransport(FTornadoTransport):
    """FNatsTransport is an extension of FTransport. This is a "stateless"
    transport in the sense that there is no connection with a server. A request
    is simply published to a subject and responses are received on another
    subject. This assumes requests/responses fit within a single NATS message.
    """

    def __init__(self, nats_client, subject, inbox="", is_ttransport=False):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        super(FNatsTransport, self).__init__()
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub_id = None

        # TODO: Remove with 2.0
        self._open_lock = locks.Lock()
        self._is_ttransport = is_ttransport

    @gen.coroutine
    def isOpen(self):
        with (yield self._open_lock.acquire()):
            # Tornado requires we raise a special exception to return a value.
            raise gen.Return(
                self._is_open and self._nats_client.is_connected())

    @gen.coroutine
    def open(self):
        """Subscribes to the configured inbox subject"""
        if not self._nats_client.is_connected():
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "NATS not connected.")

        elif (yield self.isOpen()):
            raise TTransportException(TTransportException.ALREADY_OPEN,
                                      "NATS transport already open")

        handler = self._on_message_callback
        if self._is_ttransport:
            handler = self._ttransport_on_message_callback

        self._sub_id = yield self._nats_client.subscribe(
            self._inbox,
            "",
            handler,
        )
        self._is_open = True

    def _on_message_callback(self, msg):
        self.execute_frame(msg.data)

    def _ttransport_on_message_callback(self, msg):
        if not self._execute:
            raise FExecuteCallbackNotSet("Execute callback not set")

        self._execute(msg.data)

    @gen.coroutine
    def close(self):
        """Unsubscribes from the inbox subject"""
        if not self._sub_id:
            return

        yield self._nats_client.unsubscribe(self._sub_id)
        with (yield self._open_lock.acquire()):
            self._is_open = False

    @gen.coroutine
    def flush(self):
        """Sends the buffered bytes over NATS"""
        if not (yield self.isOpen()):
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Nats not connected!")

        frame = self.get_write_bytes()
        if not frame:
            return

        self.reset_write_buffer()
        yield self._nats_client.publish_request(self._subject, self._inbox,
                                                frame)
