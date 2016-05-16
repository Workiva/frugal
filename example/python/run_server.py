import logging
import sys
sys.path.append('gen-py.tornado')

from thrift.protocol import TBinaryProtocol

from tornado import gen

from nats.io.client import Client as NATS

from frugal.processor.processor_factory import FProcessorFactory
from frugal.protocol import FProtocolFactory
from frugal.server.nats_server import FNatsServer
from frugal.transport.nats_service_transport import FNatsServiceTransportFactory

from gen_py.example.f_foo import Processor as FFooProcessor


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


# TODO: implement FFoo.Iface
class FooHandler(object):

    def ping(self, ctx):
        print "Received ping with cid : {}".format(ctx.get_corr_id())

    def oneWay(self, ctx, req):
        print "Received oneWay: {} {}".format(ctx, req)

    def blah(self, ctx, num, Str, event):
        print "Received blah {} {} {} {}".format(ctx, num, Str, event)

    def basePing(self, ctx):
        print "Received basePing: {}".format(ctx)


@gen.coroutine
def main():

    nats_client = NATS()
    options = {
        "verbose": True,
        "servers": ["nats://127.0.0.1:4222"]
    }

    yield nats_client.connect(**options)

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())
    transport_factory = FNatsServiceTransportFactory(nats_client)

    handler = FooHandler()
    processor = FFooProcessor(handler)
    processor_factory = FProcessorFactory(processor)

    subject = "foo"
    heartbeat_interval = 20 * 1000
    max_missed_heartbeats = 3

    server = FNatsServer(nats_client,
                         subject,
                         heartbeat_interval,
                         max_missed_heartbeats,
                         processor_factory,
                         transport_factory,
                         prot_factory)

    logging.info("Starting server...")

    # This should start the ioloop
    server.serve()

if __name__ == '__main__':
    main()
