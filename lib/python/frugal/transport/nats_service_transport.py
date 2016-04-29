import json
import logging
import base64
from datetime import timedelta
import struct
from threading import Lock
from io import BytesIO

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen
from tornado import concurrent
from tornado import ioloop


_NATS_PROTOCOL_VERSION = 0
_NATS_MAX_MESSAGE_SIZE = 1048576
_FRUGAL_PREFIX = "frugal."
_DISCONNECT = "DISCONNECT"
_HEARTBEAT_GRACE_PERIOD = 50000
_HEARTBEAT_LOCK = Lock()
DEFAULT_CONNECTION_TIMEOUT = 20000
DEFAULT_MAX_MISSED_HEARTBEATS = 3

logger = logging.getLogger(__name__)


class TNatsServiceTransport(TTransportBase):

    def __init__(self,
                 nats_client,
                 connection_subject,
                 connection_timeout=DEFAULT_CONNECTION_TIMEOUT,
                 max_missed_heartbeats=DEFAULT_MAX_MISSED_HEARTBEATS,
                 io_loop=None):
        """Create a TNatsServerTransport to communicate with NATS

        Args:
            connection_subject: nats connection subject
            listen_to: subject to listen on
            write_to: subject to write to
        """
        self._nats_client = nats_client
        self._io_loop = io_loop or ioloop.IOLoop.current()

        self._connection_subject = connection_subject
        self._connection_timeout = connection_timeout
        self._max_missed_heartbeats = max_missed_heartbeats

        self._is_open = False

        self._missed_heartbeats = 0

        self._open_lock = Lock()
        self._wbuf = BytesIO()

    def set_execute_callback(self, execute):
        self._execute = execute

    def isOpen(self):
        with self._open_lock:
            return self._is_open and self._nats_client.is_connected()

    @gen.coroutine
    def open(self):
        """Open the Transport to communicate with NATS

        Throws:
            TTransportException
        """
        if not self._nats_client.is_connected():
            ex = TTransportException(TTransportException.NOT_OPEN,
                                     "NATS not connected.")
            logger.error(ex.message)
            raise ex

        elif self.isOpen():
            ex = TTransportException(TTransportException.ALREADY_OPEN,
                                     "NATS transport already open")
            logger.error(ex.message)
            raise ex

        with self._open_lock:
            if self._connection_subject:
                yield self._handshake()

            self._sub_id = yield self._nats_client.subscribe(
                self._listen_to,
                "",
                self._on_message_callback
            )

            yield self._setup_heartbeat()
            self._is_open = True
            logger.info("frugal: transport open.")

    def _on_message_callback(self, msg=None):
        if msg.reply == _DISCONNECT:
            logger.debug("Received DISCONNECT from Frugal server.")
            self.close()
        else:
            self._execute(msg.data)

    @gen.coroutine
    def _handshake(self):
        inbox = self._new_frugal_inbox()
        handshake = json.dumps({"version": _NATS_PROTOCOL_VERSION})

        future = concurrent.Future()
        sid = yield self._nats_client.subscribe(inbox, b'', None, future)
        yield self._nats_client.auto_unsubscribe(sid, 1)
        yield self._nats_client.publish_request(self._connection_subject,
                                                inbox,
                                                handshake)
        # TODO replace hardcoded time
        msg = yield gen.with_timeout(
            timedelta(milliseconds=30000), future)

        subjects = msg.data.split()
        if len(subjects) != 3:
            logger.error("Bad Frugal handshake")
            # TODO handle similar to other libraries
            return
        self._heartbeat_listen = subjects[0]
        self._heartbeat_reply = subjects[1]
        self._heartbeat_interval = int(subjects[2])

        # TODO make sure listen to isn't null or empty
        self._listen_to = msg.subject
        self._write_to = msg.reply

    def _on_heartbeat_message(self, msg=None):
        logger.debug("Received heartbeat.")
        self._heartbeat_timer.stop()
        self._nats_client.publish(self._heartbeat_reply, "")
        self._missed_heartbeats = 0
        self._heartbeat_timer.start()

    @gen.coroutine
    def _setup_heartbeat(self):
        if self._heartbeat_interval > 0:
            logger.debug("Setting up heartbeat subscription.")
            self._heartbeat_sub_id = yield self._nats_client.subscribe(
                self._heartbeat_listen,
                "",
                self._on_heartbeat_message
            )

        self._heartbeat_timer = ioloop.PeriodicCallback(
            self._missed_heartbeat,
            self._heartbeat_interval
        )
        logger.debug("Starting heartbeat timer.")
        self._heartbeat_timer.start()

    @gen.coroutine
    def _missed_heartbeat(self, future=None):
        self._missed_heartbeats += 1
        if self._missed_heartbeats >= self._max_missed_heartbeats:
            logger.error("Missed maximum number ({})of heartbeats." +
                         "Closing NATS transport"
                         .format(self._missed_heartbeats))
            yield self.close()
            self._heartbeat_timer.stop()

    @gen.coroutine
    def close(self):
        """Close the transport asynchronously"""

        if self._is_open:
            # TODO check close callback
            # unsub from heartbeat

            yield self._nats_client.close()

    def read(self, buff, offset, length):
        raise Exception("Don't call this.")

    def write(self, buff):
        """Write takes in a bytearray and appends it to the write buffer"""
        if not self.isOpen():
            logger.error("Tried to write to closed transport!")
            raise TTransportException(TTransportException.NOT_OPEN,
                                      "Transport not open!")
        self._wbuf.write(buff)

    @gen.coroutine
    def flush(self):
        """Flush publishes whatever is in the buffer to NATS"""
        frame = self._wbuf.getvalue()
        frame_length = struct.pack('>I', len(frame))
        self._wbuf = BytesIO()
        yield self._nats_client.publish(self._write_to,
                                        frame_length + frame)

    def _new_frugal_inbox(self):
        return "{frugal}{new_inbox}".format(frugal=_FRUGAL_PREFIX,
                                            new_inbox=new_inbox())

