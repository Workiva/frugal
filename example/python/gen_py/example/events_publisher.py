from thrift.Thrift import TMessageType
from tornado import gen

_DELIMETER = "."


class EventsPublisher(object):


    def __init__(self, scope_provider):
        """Creates an instance of EventsPublisher

        Args:
            scope_provider: FScopeProvider
        """
        self._scope_provider = scope_provider

    @gen.coroutine
    def open(self):
        (trans, prot) = self._scope_provider.new()
        self._trans = trans
        self._prot = prot
        yield self._trans.open()

    @gen.coroutine
    def close(self):
        yield self._trans.close()

    @gen.coroutine
    def publish_event_created(self, ctx, user, req):
        op = "EventCreated"
        prefix = "foo.{}".format(user)
        topic = "{prefix}{delimeter}Event{delimeter}{op}".format(
            prefix=prefix,
            delimeter=_DELIMETER,
            op=op
        )
        print("locking topic {}".format(topic))
        self._trans.lock_topic(topic)
        try:
            self._prot.write_request_headers(ctx)
            self._prot.writeMessageBegin(op, TMessageType.CALL, 0)
            req.write(self._prot)
            self._prot.writeMessageEnd()
            self._trans.flush()
        finally:
            self._trans.unlock_topic()
