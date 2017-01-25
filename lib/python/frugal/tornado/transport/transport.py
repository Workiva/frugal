from tornado import gen

from frugal.transport import FTransport


class FTransportBase(FTransport):
    """
    FTransportBase extends FTransport using the coroutine decorators used by
    all tornado FTransports.
    """
    def is_open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def open(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def close(self):
        raise NotImplementedError("You must override this.")

    @gen.coroutine
    def oneway(self, context, payload):
        raise NotImplementedError('You must override this.')

    @gen.coroutine
    def request(self, context, payload):
        raise NotImplementedError('You must override this.')

