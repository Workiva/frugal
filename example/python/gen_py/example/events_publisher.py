from thrift.Thrift import TMessageType
from tornado import gen


class EventsPublisher(object):

    _DELIMETER = "."

    def __init__(self, provider):
        """Creates an instance of EventsPublisher

        Args:
            scope_provider: FScopeProvider
        """
        self._transport, protocol_factory = provider.new()
        self._protocol = protocol_factory.get_protocol(self._transport)

    @gen.coroutine
    def open(self):
        yield self._transport.open()

    @gen.coroutine
    def close(self):
        yield self._transport.close()

    @gen.coroutine
    def publish_event_created(self, ctx, user, req):
        op = "EventCreated"
        prefix = "foo.{}".format(user)
        topic = "{prefix}{delimeter}Events{delimeter}{op}".format(
            prefix=prefix,
            delimeter=self._DELIMETER,
            op=op
        )
        self._transport.lock_topic(topic)
        try:
            self._protocol.write_request_headers(ctx)
            self._protocol.writeMessageBegin(op, TMessageType.CALL, 0)
            print "BEFORE WRITE {} {}".format(req.ID, req.Message)
            req.write(self._protocol)
            self._protocol.writeMessageEnd()
            self._transport.flush()
        finally:
            self._transport.unlock_topic()
