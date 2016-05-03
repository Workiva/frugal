import logging

from .server import FServer

logger = logging.getLogger(__name__)

# TODO: implement this.
DEFAULT_MAX_MISSED_HEARTBEATS = 2


class FNatsServer(FServer):

    def __init__(self,
                 nats_client,
                 subject,
                 heartbeat_interval,
                 max_missed_heartbeats=DEFAULT_MAX_MISSED_HEARTBEATS):
        self._nats_client = nats_client
        self._subject = subject
        self._heartbeat_interval = heartbeat_interval
        self._max_missed_heartbeats = max_missed_heartbeats

    def serve(self):
        pass
        # self.connection.QueueSubscribe(self.subject, queue, callback)

