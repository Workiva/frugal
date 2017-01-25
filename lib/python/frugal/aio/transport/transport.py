from thrift.transport.TTransport import TTransportBase

from frugal.transport import FTransport


class FTransportBase(FTransport):
    """
    FBaseTransport extends FTransport using the async decorators used by all
    asyncio FTransports.
    """
    def is_open(self) -> bool:
        raise NotImplementedError('You must override this')

    async def open(self):
        raise NotImplementedError('You must override this')

    async def close(self):
        raise NotImplementedError('You must override this')

    async def oneway(self, context, payload):
        raise NotImplementedError('You must override this')

    async def request(self, context, payload) -> TTransportBase:
        raise NotImplementedError('You must override this')

