#
# Autogenerated by Frugal Compiler (1.9.2)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



import sys
import traceback

from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from thrift.Thrift import TType
from tornado import gen
from frugal.middleware import Method
from frugal.subscription import FSubscription

from valid.ttypes import *




class blahSubscriber(object):

    _DELIMITER = '.'

    def __init__(self, provider, middleware=None):
        """
        Create a new blahSubscriber.

        Args:
            provider: FScopeProvider
            middleware: ServiceMiddleware or list of ServiceMiddleware
        """

        if middleware and not isinstance(middleware, list):
            middleware = [middleware]
        self._middleware = middleware
        self._transport, self._protocol_factory = provider.new()

    @gen.coroutine
    def subscribe_DoStuff(self, DoStuff_handler):
        """
            DoStuff_handler: function which takes FContext and Thing
        """

        op = 'DoStuff'
        prefix = ''
        topic = '{}blah{}{}'.format(prefix, self._DELIMITER, op)

        yield self._transport.subscribe(topic, self._recv_DoStuff(self._protocol_factory, op, DoStuff_handler))

    def _recv_DoStuff(self, protocol_factory, op, handler):
        method = Method(handler, self._middleware)

        def callback(transport):
            iprot = protocol_factory.get_protocol(transport)
            ctx = iprot.read_request_headers()
            mname, _, _ = iprot.readMessageBegin()
            if mname != op:
                iprot.skip(TType.STRUCT)
                iprot.readMessageEnd()
                raise TApplicationException(TApplicationException.UNKNOWN_METHOD)
            req = Thing()
            req.read(iprot)
            iprot.readMessageEnd()
            try:
                method([ctx, req])
            except:
                traceback.print_exc()
                sys.exit(1)

        return callback




