import json
import logging
from datetime import timedelta

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen
from tornado import concurrent
from tornado import ioloop


_NATS_MAX_MESSAGE_SIZE = 1048576
_FRUGAL_PREFIX = "frugal."
_DISCONNECT = "DISCONNECT"
_HEARTBEAT_GRACE_PERIOD = 50000

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

    def is_open(self):
        return self._is_open and self._nats_client.is_connected()

    @gen.coroutine
    def open(self):
        """Open the Transport

        Throws:
            TTransportException
        """

        if not self._nats_client.is_connected():
            raise TTransportException(1, "NATS not connected.")
        elif self.is_open():
            raise TTransportException(2, "NATS transport already open")

        if self._connection_subject:
            msg = yield self._handshake()
            print("got handshake message with subject: {0} and data: {1}".format(msg.subject, msg.data))

        def on_message_cb(msg=None):
            if msg.reply == _DISCONNECT:
                self.close()
                return
            # TODO write msg.data to writer
            print("message subject: {0}, data: {1}".format(msg.subject, msg.data))

        self._sub_id = yield self._nats_client.subscribe(self._listen_to, "", on_message_cb)
        print("self sub id = {}".format(self._sub_id))

        def on_heartbeat_message_cb(msg=None):
            print("heartbeat callback called")
            self._heartbeat_timer = ioloop.PeriodicCallback(
                self._send_ping,
                self._heartbeat_interval
            )
            self._heartbeat_timer.start()

        if self._heartbeat_interval > 0:
            self._heartbeat_sub_id = yield self._nats_client.subscribe(
                self._heartbeat_listen,
                "",
                on_heartbeat_message_cb
            )

        self._is_open = True

        raise gen.Return(self)

    @gen.coroutine
    def _handshake(self):
        inbox = self._new_frugal_inbox()
        handshake = json.dumps({"version": 0})

        future = concurrent.Future()
        sid = yield self._nats_client.subscribe(inbox, b'', None, future)
        yield self._nats_client.auto_unsubscribe(sid, 1)
        yield self._nats_client.publish_request(self._connection_subject,
                                                inbox,
                                                handshake)
        # TODO replace hardcoded time
        msg = yield gen.with_timeout(timedelta(milliseconds=50000), future)

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

        raise gen.Return(msg)

    @gen.coroutine
    def _send_ping(self, future=None):
        if self._pings_outstanding > self.options["max_outstanding_pings"]:
            yield self._unbind()
        else:
            yield self._nats_client.publish(self._heartbeat_reply, None)
            if future is None:
                future = concurrent.Future()
        self._pings_outstanding += 1
        self._pongs.append(future)

    @gen.coroutine
    def _unbind(self):
        if (self._nats_client.is_connecting() or
                self._nats_client.is_closed() or
                self._nats_client.is_reconnecting()):
            return
        if self._disconnected_cb is not None:
            self._disconnected_cb()

        yield self.close()

    @gen.coroutine
    def close(self):
        """Close the transport asynchronously"""

        if self._is_open:
            # TODO check close callback
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
