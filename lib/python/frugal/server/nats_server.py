import json
import logging
import re
from threading import Lock

from nats.io.utils import new_inbox
from tornado import gen, ioloop

from frugal.server import FServer
from frugal.transport import FTransport, TNatsServiceTransport
from frugal.registry import FServerRegistry

logger = logging.getLogger(__name__)

_NATS_PROTOCOL_V0 = 0
_DEFAULT_MAX_MISSED_HEARTBEATS = 2
_DEFAULT_HEARTBEAT_INTERVAL = 20000
_QUEUE = "rpc"


class FNatsTornadoServer(FServer):

    def __init__(self,
                 nats_client,
                 subject,
                 max_missed_heartbeats,
                 processor_factory,
                 transport_factory,
                 protocol_factory,
                 heartbeat_interval=_DEFAULT_HEARTBEAT_INTERVAL,
                 high_watermark=FTransport.DEFAULT_HIGH_WATERMARK):
        """Create a new instance of FNatsTornadoServer

        Args:
            nats_client: connected instance of nats.io.Client
            subject: subject to listen on
            heartbeat_interval: how often to send heartbeats in millis
            max_missed_heartbeats: number of heartbeats client can miss
            processor_factory: FProcessFactory
            tranpsort_factory: FTransportFactory
            protocol_factory: FProtocolFactory
        """
        self._nats_client = nats_client
        self._subject = subject
        self._heartbeat_subject = new_inbox()
        self._heartbeat_interval = heartbeat_interval
        self._max_missed_heartbeats = max_missed_heartbeats
        self._processor_factory = processor_factory
        self._transport_factory = transport_factory
        self._protocol_factory = protocol_factory
        self._high_watermark = high_watermark
        self._clients_lock = Lock()
        self._clients = {}

    @gen.coroutine
    def serve(self):
        """Subscribe to provided subject and listen on "rpc" queue."""
        logger.debug("Starting Frugal NATS Server...")

        self._sid = yield self._nats_client.subscribe(
            self._subject,
            _QUEUE,
            self._on_message_callback
        )

        if self._heartbeat_interval > 0:
            self._heartbeater = ioloop.PeriodicCallback(
                self._send_heartbeat,
                self._heartbeat_interval
            )
            self._heartbeater.start()

    @gen.coroutine
    def stop(self):
        """Stop listening for RPC calls."""
        logger.debug("Shutting down Frugal NATS Server.")
        with self._clients_lock:
            for _, client in self._clients.iteritems():
                yield client.kill()
            self._clients.clear()

        if self._heartbeater.is_running():
            self._heartbeater.stop()

    def set_high_watermark(self, high_watermark):
        """Set the high watermark value for the server

        Args:
            watermark: long representing high watermark value
        """
        self._high_watermark = high_watermark

    def get_high_watermark(self):
        return self._high_watermark

    def _new_frugal_inbox(self, prefix):
        tokens = re.split('\.', prefix)
        tokens[len(tokens) - 1] = new_inbox()
        inbox = ""
        pre = ""
        for token in tokens:
            inbox += pre + token
            pre = "."
        return inbox

    @gen.coroutine
    def _accept(self, listen_to, reply_to, heartbeat_subject):
        client = TNatsServiceTransport.Server(
            self._nats_client,
            listen_to,
            reply_to
        )
        transport = self._transport_factory.get_transport(client)
        processor = self._processor_factory.get_processor(transport)
        protocol = self._protocol_factory.get_protocol(transport)
        transport.set_registry(
            FServerRegistry(processor, self._protocol_factory, protocol)
        )
        yield transport.open()
        raise gen.Return(client)

    @gen.coroutine
    def _remove(self, heartbeat):
        with self._clients_lock:
            client = self._clients.pop(heartbeat, None)
        if client:
            yield client.kill()

    @gen.coroutine
    def _send_heartbeat(self):
        with self._clients_lock:
            if len(self._clients) == 0:
                return
        yield self._nats_client.publish(self._heartbeat_subject, "")

    @gen.coroutine
    def _on_message_callback(self, msg):
        reply_to = msg.reply
        if not reply_to:
            logger.warn("Received a bad connection handshake. Discarding.")
            return

        conn_protocol = json.loads(msg.data)
        version = conn_protocol.get('version')
        if version != _NATS_PROTOCOL_V0:
            logger.error("Version {} is not a supported NATS connect version"
                         .format(version))

        heartbeat = new_inbox()
        listen_to = self._new_frugal_inbox(msg.reply)

        transport = yield self._accept(listen_to, reply_to, heartbeat)

        client = _Client(self._nats_client,
                         transport,
                         heartbeat,
                         self._max_missed_heartbeats,
                         self._heartbeat_interval)

        if self._heartbeat_interval > 0:
            client.start()
            with self._clients_lock:
                self._clients[heartbeat] = client

        # [heartbeat_subject] [heartbeat_reply] [heartbeat_interval]
        connect_msg = "{0} {1} {2}".format(
            self._heartbeat_subject,
            heartbeat,
            self._heartbeat_interval
        )

        # TODO: Handle Exceptions
        yield self._nats_client.publish_request(
            reply_to,
            listen_to,
            connect_msg
        )


class _Client(object):

    def __init__(self,
                 nats_client,
                 transport,
                 heartbeat,
                 heartbeat_interval,
                 max_missed_heartbeats):
        self._nats_client = nats_client
        self._transport = transport
        self._heartbeat = heartbeat
        self._heartbeat_interval = heartbeat_interval
        self._max_missed_heartbeats = max_missed_heartbeats
        self._missed_heartbeats = 0
        self._heartbeat_lock = Lock()

    @gen.coroutine
    def start(self):
        self._hb_sub_id = yield self._nats_client.subscribe(
            self._heartbeat,
            "",
            self._receive_heartbeat)
        # start the timer
        self._heartbeat_timer = ioloop.PeriodicCallback(
            self._missed_heartbeat,
            self._heartbeat_interval
        )

    def _receive_heartbeat(self, msg):
        with self._heartbeat_lock:
            self._missed_heartbeats = 0

    @gen.coroutine
    def _missed_heartbeat(self, msg):
        with self._heartbeat_lock:
            self._missed_heartbeats += 1
            if self._missed_heartbeats > self._max_missed_heartbeats:
                logger.warn("Client heartbeat expired.")
                yield self.kill()

    @gen.coroutine
    def kill(self):
        logger.debug("Client disconnected.")
        if (hasattr(self, '_heartbeat_timer') and
                self._heartbeat_timer.is_running()):
            self._heartbeat_timer.stop()
        yield self._nats_client.unsubscribe(self._hb_sub_id)
        yield self._transport.close()

