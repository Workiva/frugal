from gnats.client.errors import ErrConnectionClosed, ErrMaxPayload
from gnats.client.utils import new_inbox

from thrift.transport.TTransport import TTransportException

from frugal.gevent.transport import FGeventTransport

_NOT_OPEN = 'NATS not connected.'
_ALREAD_OPEN = 'NATS transport already open.'


class FNatsTransport(FGeventTransport):
    """FNatsTransport is an extension of FTransport. This is a "stateless"
    transport in the sense that there is no connection with a server. A request
    is simply published to a subject and responses are received on another
    subject. This assumes requests/responses fit within a single NATS message.
    """

    def __init__(self, nats_client, subject, inbox=""):
        """Create a new instance of FNatsTransport

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to publish to
        """
        super(FNatsTransport, self).__init__()
        self._nats_client = nats_client
        self._subject = subject
        self._inbox = inbox or new_inbox()
        self._is_open = False
        self._sub = None

    def is_open(self):
        return self._is_open and self._nats_client.is_connected

    def open(self):
        """Subscribes to the configured inbox subject"""
        if not self._nats_client.is_connected:
            raise TTransportException(TTransportException.NOT_OPEN, _NOT_OPEN)

        elif self.is_open():
            already_open = TTransportException.ALREADY_OPEN
            raise TTransportException(already_open, _ALREAD_OPEN)

        cb = self._on_message_callback
        inbox = self._inbox
        self._sub = self._nats_client.subscribe(inbox, cb=cb)

        self._is_open = True

    def _on_message_callback(self, msg):
        self.execute_frame(msg.data)

    def close(self):
        """Unsubscribes from the inbox subject"""
        if not self._sub:
            return
        self._nats_client.flush()
        self._nats_client.unsubscribe(self._sub)
        self._is_open = False

    def send(self, data):
        """Sends the buffered bytes over NATS"""
        if not self.is_open():
            raise TTransportException(TTransportException.NOT_OPEN, _NOT_OPEN)

        subject = self._subject
        inbox = self._inbox

        try:
            self._nats_client.publish(subject, data, reply=inbox)
        except (ErrMaxPayload, ErrConnectionClosed) as e:
            raise TTransportException(
                TTransportException.UNKNOWN,
                'Error publishing to nats: {e}'.format(e=e))
