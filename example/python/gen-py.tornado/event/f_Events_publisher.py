#
# Autogenerated by Frugal Compiler (1.3.0)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from thrift.Thrift import TType
from tornado import gen
from frugal.subscription import FSubscription

from event.ttypes import *




class EventsPublisher(object):
    """
    This docstring gets added to the generated code because it has
    the @ sign. Prefix specifies topic prefix tokens, which can be static or
    variable.
    """

    _DELIMITER = '.'

    def __init__(self, provider):
        """
        Create a new EventsPublisher.

        Args:
            provider: FScopeProvider
        """

        self._transport, protocol_factory = provider.new()
        self._protocol = protocol_factory.get_protocol(self._transport)

    def open(self):
        self._transport.open()

    def close(self):
        self._transport.close()

    def publish_EventCreated(self, ctx, user, req):
        """
        This is a docstring.
        
        Args:
            ctx: FContext
            user: string
            req: Event
        """

        op = 'EventCreated'
        prefix = 'foo.%s.' % (user)
        topic = '%sEvents%s%s' % (prefix, self._DELIMITER, op)
        oprot = self._protocol
        self._transport.lock_topic(topic)
        try:
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(op, TMessageType.CALL, 0)
            req.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()
        finally:
            self._transport.unlock_topic()

