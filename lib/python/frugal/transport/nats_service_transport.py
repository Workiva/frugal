import json
import logging
from datetime import timedelta
from threading import Lock

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
                 max_missed_heartbeats=DEFAULT_MAX_MISSED_HEARTBEATS):
        """Create a TNatsServerTransport to communicate with NATS

        Args:
            connection: nats connection
            listen_to: subject to listen on
            write_to: subject to write to
        """
        # self.io_loop = io_loop or ioloop.IOLoop.current()

        self._nats_client = nats_client

        self._connection_subject = connection_subject
        self._connection_timeout = connection_timeout
        self._max_missed_heartbeats = max_missed_heartbeats

        self._is_open = False
        self._sub_id = None

        self._heartbeat_listen = None
        self._heartbeat_reply = None
        self._heartbeat_interval = None
        self._heartbeat_timer = None
        self._heartbeat_sub_id = None

        self._write_to = None
        self._listen_to = None
        self._heartbeat_count = 0

    def isOpen(self):
        return self._is_open and self._nats_client.is_connected()

    @gen.coroutine
    def open(self):
        """Open the Transport

        Throws:
            TTransportException
        """

        #
        # handle exceptions
        #
        if not self._nats_client.is_connected():
            raise TTransportException(1, "NATS not connected.")
        elif self.isOpen():
            raise TTransportException(2, "NATS transport already open")

        #
        # handshake
        #
        if self._connection_subject:
            yield self._handshake()

        #
        # subscribe to topic
        #
        def on_message_cb(msg=None):
            if msg.reply == _DISCONNECT:
                self.close()
                return
            # TODO write msg.data to writer
            print("subject: {0}, data: {1}".format(msg.subject, msg.data))

        self._sub_id = yield self._nats_client.subscribe(
            self._listen_to,
            "",
            on_message_cb
        )

        self._setup_heartbeat()

        self._is_open = True

        # raise gen.Return()

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

        raise gen.Return(msg)
        print("message data: {}".format(msg.data))
        subjects = msg.data.split()
        if len(subjects) != 3:
            print("bad handshake")
        self._heartbeat_listen = subjects[0]
        self._heartbeat_reply = subjects[1]
        self._heartbeat_interval = int(subjects[2])

        # TODO make sure listen to isn't null or empty
        self._listen_to = msg.subject
        self._write_to = msg.reply

    @gen.coroutine
    def _setup_heartbeat(self):
        def on_heartbeat_message_cb(msg=None):
            print("message subject {0}, data {1}".format(msg.subject, msg.data))
            with _HEARTBEAT_LOCK:
                self._heartbeat_count += 1
                print("heartbeat count: {}".format(self._heartbeat_count))
                self._heartbeat_timer.stop()
                self._missed_heartbeats = 0
                print("heartbeat callback called missed heartbeats: {} should be 0 ".format(self._missed_heartbeats))
                self._heartbeat_timer.start()

        if self._heartbeat_interval > 0:
            self._heartbeat_sub_id = yield self._nats_client.subscribe(
                self._heartbeat_listen,
                "",
                on_heartbeat_message_cb
            )

        self._heartbeat_timer = ioloop.PeriodicCallback(
            self._missed_heartbeat,
            self._heartbeat_interval
        )
        self._heartbeat_timer.start()

    @gen.coroutine
    def _missed_heartbeat(self, future=None):
        with _HEARTBEAT_LOCK:
            self._missed_heartbeats += 1
            print("missed heartbeats {}".format(self._missed_heartbeats))
            if self._missed_heartbeats >= self._max_missed_heartbeats:
                print("Exceeded maximum number of acceptable " +
                    "missed heartbeats.  Closing transport.")
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
        pass

    def write(self, buff, offset, length):
        pass

    def flush(self):
        """flush publishes whatever is in the buffer to NATS"""
        pass

    def _new_frugal_inbox(self):
        return "{0}{1}".format(_FRUGAL_PREFIX, new_inbox())
