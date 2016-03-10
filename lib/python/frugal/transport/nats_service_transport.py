import json
from datetime import timedelta

from nats.io.utils import new_inbox
from thrift.transport.TTransport import TTransportBase, TTransportException
from tornado import gen
from tornado.concurrent import Future

_NATS_MAX_MESSAGE_SIZE = 1048576
_FRUGAL_PREFIX = "frugal."
_DISCONNECT = "DISCONNECT"
_HEARTBEAT_GRACE_PERIOD = 50000


class TNatsServiceTransport(TTransportBase):

    def __init__(self, nats_client, connection_subject,
                 connection_timeout, max_missed_heartbeats):
        """Create a TNatsServerTransport to communicate with NATS

        Args:
            connection: nats connection
            listen_to: subject to listen on
            write_to: subject to write to
        """

        self._nats_client = nats_client
        self._connection_subject = connection_subject
        self._connection_timeout = connection_timeout
        self._max_missed_heartbeats = max_missed_heartbeats

    def is_open(self):
        #TODO: fix this
        return self._is_open

    @gen.coroutine
    def open(self):
        """Open the Transport

            Throws:
                TTransportException
        """
        if not self._nats_client.is_connected():
            raise TTransportException(
                1,
                "NATS not connected.")
        elif self.isOpen():
            raise TTransportException(
                2,
                "NATS transport already open")

        inbox = self._new_frugal_inbox()

        conn_handshake = {"version": 0}
        encoded_handshake = json.dumps(conn_handshake)

        future = Future()
        sid = yield self._nats_client.subscribe(inbox, b'', None, future)
        yield self._nats_client.auto_unsubscribe(sid, 1)
        yield self._nats_client.publish_request(self._connection_subject,
                                                inbox,
                                                encoded_handshake)

        msg = yield gen.with_timeout(timedelta(milliseconds=5000), future)
        subjects = msg.data.split()
        print(str(subjects))
        self._heartbeat_listen = subjects[0]
        self._heartbeat_reply = subjects[1]
        self._heartbeat_interval = int(subjects[2])
        self._listen_to = msg.subject
        self.reply = msg.reply

        # raise gen.Return(self)

    @gen.coroutine
    def close(self):
        """Close the transport asynchronously"""

        if self._is_open:
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


    def _receive_heartbeat(self):
        pass

    def _missed_heartbeat(self):
        pass

    def _start_timer(self):
        pass
