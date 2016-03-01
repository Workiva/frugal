from threading import Lock

from thrift.transport.TTransport import TTransportBase, TTransportException


_NATS_MAX_MESSAGE_SIZE = 1048576
_FRUGAL_PREFIX = "frugal."
_DISCONNECT = "DISCONNECT"
_HEARTBEAT_GRACE_PERIOD = 50000


class TNatsServiceTransport(TTransportBase):

    def __init__(self, conn, listen_to, write_to):
        """Create a TNatsServerTransport to communicate with NATS

        Args:
            connection: nats connection
            listen_to: subject to listen on
            write_to: subject to write to
        """

        self._conn = conn
        self._listen_to = listen_to
        self._write_to = write_to
        self._is_open = False

    def open(self):
        """Open the Transport

            Throws:
                TTransportException
        """
        if not self._conn.is_connected():
            raise TTransportException(
                1,
                "NATS not connected.")
        elif self.isOpen():
            raise TTransportException(
                2,
                "NATS transport already open")
        else:
            print "do some stuff"

    def close(self):
        """Close the transport"""

        if self._is_open:
            # try publish disconnect

            self._is_open = False

    def read(self, buff, offset, length):
        pass

    def write(self, buff, offset, length):
        pass

    def flush(self):
        pass

    def _handshake(self):
        pass

    def _receive_heartbeat(self):
        pass

    def _missed_heartbeat(self):
        pass

    def _start_timer(self):
        pass
