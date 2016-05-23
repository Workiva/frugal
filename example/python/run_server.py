import logging
import sys
sys.path.append('gen-py.tornado')

from thrift.protocol import TBinaryProtocol

from tornado import gen, ioloop

from nats.io.client import Client as NATS

from frugal.processor.processor_factory import FProcessorFactory
from frugal.protocol.protocol_factory import FProtocolFactory
from frugal.server import FNatsTornadoServer
from frugal.transport import FMuxTornadoTransportFactory

from event.f_Foo import Iface, Processor as FFooProcessor


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


class ExampleHandler(Iface):

    def ping(self, ctx):
        print "ping: {}".format(ctx)

    def oneWay(self, ctx, id, req):
        print "oneWay: {} {} {}".format(ctx, id, req)

    def blah(self, ctx, num, Str, event):
        print "blah: {} {} {} {}".format(ctx, num, Str, event)
        ctx.set_response_header("foo", "bar")
        return 42

    def basePing(self, ctx):
        print "basePing: {}".format(ctx)


@gen.coroutine
def main():

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    transport_factory = FMuxTornadoTransportFactory()

    handler = ExampleHandler()
    processor = FFooProcessor(handler)
    processor_factory = FProcessorFactory(processor)

    subject = "foo"
    heartbeat_interval = 10000
    max_missed_heartbeats = 3

    server = FNatsTornadoServer(nats_client,
                                subject,
                                max_missed_heartbeats,
                                processor_factory,
                                transport_factory,
                                prot_factory,
                                heartbeat_interval)

    logging.info("Starting server...")

    yield server.serve()

if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
