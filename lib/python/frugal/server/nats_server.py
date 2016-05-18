import json
import logging
import re

from nats.io.utils import new_inbox
from tornado import gen, ioloop

from frugal.server import FServer
from frugal.transport import TNatsServiceTransport
from frugal.registry import FServerRegistry

logger = logging.getLogger(__name__)

_NATS_PROTOCOL_VERSION = 0
_DEFAULT_MAX_MISSED_HEARTBEATS = 2
_QUEUE = "rpc"


class FNatsServer(FServer):

    def __init__(self,
                 nats_client,
                 subject,
                 heartbeat_interval,
                 max_missed_heartbeats,
                 processor_factory,
                 transport_factory,
                 protocol_factory):
        """Create a new instance of FNatsServer

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
        self._heartbeat_interval = heartbeat_interval or 10000
        self._max_missed_heartbeats = max_missed_heartbeats
        self._processor_factory = processor_factory
        self._transport_factory = transport_factory
        self._protocol_factory = protocol_factory
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
        # stop the timers
        for _, client in self._clients.iteritems():
            client.kill()
        self._clients.clear()
        self._heartbeater.stop()

    def set_high_watermark(self, watermark):
        """Set the high watermark value for the server

        Args:
            watermark: long representing high watermark value
        """
        self._watermark = watermark

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

    def _remove(self, heartbeat):
        client = self._clients.pop(heartbeat, None)
        if client:
            client.kill()

    @gen.coroutine
    def _send_heartbeat(self):
        if len(self._clients) == 0:
            return
        yield self._nats_client.publish(self._heartbeat_subject, "")

    @gen.coroutine
    def _on_message_callback(self, msg=None):
        reply_to = msg.reply
        if not reply_to:
            logger.warn("Received a bad connection handshake. Discarding.")
            return

        conn_protocol = json.loads(msg.data)
        version = conn_protocol['version']
        print "version %s", version
        if version != _NATS_PROTOCOL_VERSION:
            logger.error("Version {} is not a supported NATS connect version"
                         .format(version))

        heartbeat = new_inbox()
        listen_to = self._new_frugal_inbox(msg.reply)

        transport = yield self._accept(listen_to, reply_to, heartbeat)

        client = self._Client(transport, heartbeat)

        if self._heartbeat_interval > 0:
            client.start()
            self._clients[heartbeat] = client

        # Publish back connect message [heartbeat_subject] [heartbeat_reply]
        # [heartbeat_interval]
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

        def __init__(self, transport, heartbeat, io_loop=None):
            self._transport = transport
            self._heartbeat = heartbeat
            self._io_loop = io_loop or ioloop.IOLoop.current()

        @gen.coroutine
        def start(self):
            # subscribe to the client's heartbeat
            # TODO : subscribe to client heartbeat, count missed etc...
            print "CALLED START ON CLIENT"
            # yield self._nats_client.subscribe(heartbea

        def _missed_heartbeat(self, msg=None):
            pass

        @gen.coroutine
        def kill(self):
            yield self._transport.close()

