import logging

from gevent.event import Event
from thrift.Thrift import TApplicationException
from thrift.transport.TTransport import TMemoryBuffer

from frugal import _NATS_MAX_MESSAGE_SIZE
from frugal.server import FNatsBaseServer
from frugal.transport import TMemoryOutputBuffer

logger = logging.getLogger(__name__)


class FNatsServer(FNatsBaseServer):
    """An implementation of FServer which uses NATS as the underlying transport.
    Clients must connect with the FNatsTransport"""

    def __init__(self, nats_client, subjects, processor,
                 protocol_factory, queue=""):
        """Create a new instance of FNatsServer

        Args:
            nats_client: connected instance of gnats.Client
            subjects: subject or list of subjects to listen on
            processor: FProcess
            protocol_factory: FProtocolFactory
            queue: Nats queue group
        """
        self._nats_client = nats_client
        self._subjects = [subjects] if isinstance(subjects, basestring) \
            else subjects
        self._processor = processor
        self._iprot_factory = protocol_factory
        self._oprot_factory = protocol_factory
        self._queue = queue
        self._subs = []

        self.blocking_event = Event()

    def serve(self):
        """Subscribe to provided subject and listen on provided queue"""
        queue = self._queue
        cb = self._on_message_callback

        self._subs = [
            self._nats_client.subscribe(
                subject,
                queue=queue,
                cb=cb
            ) for subject in self._subjects
        ]

        logger.info("Frugal server running...")

        self.blocking_event.wait()

    def stop(self):
        """Unsubscribe from server subject"""
        logger.debug("Frugal server stopping...")
        for sid in self._sub_ids:
            self._nats_client.unsubscribe(sid)
        self.blocking_event.set()

    def _on_message_callback(self, msg):
        """Process and respond to server request on server subject

        Args:
            msg: request message published to server subject
        """
        if not self._validate_nats_message(msg):
            return

        reply_to = msg.reply

        # Read and process frame (exclude first 4 bytes which
        # represent frame size).
        iprot = self._iprot_factory.get_protocol(
            TMemoryBuffer(msg.data[4:])
        )
        otrans = TMemoryOutputBuffer(_NATS_MAX_MESSAGE_SIZE)
        oprot = self._oprot_factory.get_protocol(otrans)

        try:
            self._processor.process(iprot, oprot)
        except TApplicationException:
            # Continue so the exception is sent to the client
            pass
        except Exception:
            return

        if self._is_message_oneway(otrans):
            return

        self._nats_client.publish(reply_to, otrans.getvalue())
