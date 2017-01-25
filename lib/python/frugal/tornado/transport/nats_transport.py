from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportException
from tornado import gen

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.tornado.transport import FAsyncTransport

_NOT_OPEN = 'NATS not connected.'
_ALREAD_OPEN = 'NATS transport already open.'


class FNatsTransport(FAsyncTransport):
    """FNatsTransport is an extension of FAsyncTransport. This is a "stateless"
    transport in the sense that there is no connection with a server. A request
    is simply published to a subject and responses are received on another
    subject. This assumes requests/responses fit within a single NATS message.
    """

    def __init__(self, nats_client, subject, inbox=""):
        """Create a new instance of FStatelessNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        super(FNatsTransport, self).__init__(
            request_size_limit=_NATS_MAX_MESSAGE_SIZE)
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub_id = None

    def is_open(self):
        return self._is_open and self._nats_client.is_connected

    @gen.coroutine
    def open(self):
        """Subscribes to the configured inbox subject"""
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportException.NOT_OPEN, _NOT_OPEN)

        elif self.is_open():
            already_open = TTransportException.ALREADY_OPEN
            raise TTransportException(already_open, _ALREAD_OPEN)

        cb = self._on_message_callback
        inbox = self._inbox
        self._sub_id = yield self._nats_client.subscribe_async(inbox, cb=cb)

        self._is_open = True

    @gen.coroutine
    def _on_message_callback(self, msg):
        yield self.handle_response(msg.data[4:])

    @gen.coroutine
    def close(self):
        """Unsubscribes from the inbox subject"""
        if not self._sub_id:
            return
        yield self._nats_client.flush()
        yield self._nats_client.unsubscribe(self._sub_id)
        self._is_open = False

    @gen.coroutine
    def flush(self, payload):
        """Sends the buffered bytes over NATS"""
        subject = self._subject
        inbox = self._inbox
        yield self._nats_client.publish_request(subject, inbox, payload)
        # If we don't flush here the ioloop waits for 2 minutes before flushing
        yield self._nats_client.flush()
