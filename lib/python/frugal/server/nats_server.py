import logging

from .server import FServer

logger = logging.getLogger(__name__)

# TODO: implement this.
_DEFAULT_MAX_MISSED_HEARTBEATS = 2
_QUEUE = "rpc"


class FNatsServer(FServer):

    def __init__(self,
                 nats_client,
                 subject,
                 heartbeat_interval,
                 max_missed_heartbeats=_DEFAULT_MAX_MISSED_HEARTBEATS):
        self._nats_client = nats_client
        self._subject = subject
        self._heartbeat_interval = heartbeat_interval
        self._max_missed_heartbeats = max_missed_heartbeats

    def serve(self):
        pass

    def stop(self):
        pass

    def set_high_watermark(self):
        pass

    def get_high_watermark(self, watermark):
        pass

    def _new_frugal_inbox(self, prefix):
        pass

    def _accept(self, listen_to, reply_to, heartbeat_subject):
        pass

    def _remove(self, heartbeat):
        pass


    class _Client(object):

        def __init__(self, transport, heartbeat):
            self._transport = transport
            self._heartbeat = heartbeat

        def start(self):
            pass

        def kill(self):
            pass

