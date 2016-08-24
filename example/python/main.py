import logging
import sys
sys.path.append('gen-py')

from thrift.protocol import TBinaryProtocol

from frugal.context import FContext
from frugal.protocol import FProtocolFactory
from frugal.transport.http_transport import FHttpTransport
from frugal.server.http_server import FHttpServer

from event.f_Foo import Client as FFooClient
from event.f_Foo import Iface
from event.f_Foo import Processor as FFooProcessor
from event.ttypes import Event


root = logging.getLogger()
root.setLevel(logging.DEBUG)

ch = logging.StreamHandler(sys.stdout)
ch.setLevel(logging.DEBUG)
formatter = logging.Formatter(
    '%(asctime)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
root.addHandler(ch)


URL = 'http://localhost:8090/frugal'


def main():
    logging.info("Starting...")

    prot_factory = FProtocolFactory(TBinaryProtocol.TBinaryProtocolFactory())

    if "-server" in sys.argv:
        root.debug("Running server")
        run_server(prot_factory)
    else:
        root.debug("Running client")
        run_client(prot_factory)


def run_server(prot_factory):
    handler = ExampleHandler()
    processor = FFooProcessor(handler)
    server = FHttpServer(processor, ('', 8090), prot_factory)
    server.serve()


def run_client(prot_factory):
    transport = FHttpTransport(URL)
    transport.open()

    foo_client = FFooClient(transport, prot_factory)

    root.info('oneWay()')
    foo_client.oneWay(FContext(), 99, {99: "request"})

    root.info('basePing()')
    foo_client.basePing(FContext(timeout=5 * 1000))

    root.info('ping()')
    foo_client.ping(FContext(timeout=1000))

    ctx = FContext()
    event = Event(42, "hello world")
    root.info('blah()')
    b = foo_client.blah(ctx, 100, "awesomesauce", event)
    root.info('Blah response {}'.format(b))
    root.info('Response header foo: {}'.format(ctx.get_response_header("foo")))

    transport.close()


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


if __name__ == '__main__':
    main()

