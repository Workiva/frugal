from threading import Lock

from .scope_transport import FScopeTransport
from frugal.exceptions import FException


class FNatsScopeTransport(FScopeTransport):

    def __init__(self, conn=None):
        self._conn = conn
        self._subject = ""
        self._lock = Lock()
        self._pull = False

    def lock_topic(self, topic):
        if self._pull:
            raise FException("Subscriber cannot lock topic.")

        self._lock.acquire()
        self._subject = topic

    def unlock_topic(self):
        if self._pull:
            raise FException("Subscriber cannot unlock topic.")

        self._subject = ""
        self._lock.release()

    def subscribe(self, topic):
        self._pull = True
        self._subject = topic
        self.open()

    def open():
        # TODO
        pass
