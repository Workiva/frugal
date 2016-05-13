from thrift.Thrift import TType, TApplicationException
from tornado import gen

from .ttypes import Event
from frugal.subscription import FSubscription


class EventsSubscriber(object):

    _DELIMETER = "."

    def __init__(self, provider):
        self._transport, self._protocol_factory = provider.new()

    @gen.coroutine
    def subscribe_event_created(self, user, event_handler):
        op = "EventCreated"
        prefix = "foo.{}.".format(user)
        topic = "{}Events{}{}".format(prefix, self._DELIMETER, op)

        yield self._transport.subscribe(
            topic,
            self.recv_EventCreated(self._protocol_factory, op, event_handler)
        )

    def recv_EventCreated(self, protocol_factory, op, event_handler):
        def event_created_callback(transport):
            iprot = protocol_factory.get_protocol(transport)
            try:
                context = iprot.read_request_headers()
                (mname, mtype, mid) = iprot.readMessageBegin()
                print("mname, type, mid {} {} {}".format(mname, mtype, mid))
                if mname != op:
                    iprot.skip(TType.STRUCT)
                    iprot.readMessageEnd()
                    raise TApplicationException(
                        TApplicationException.UNKNOWN_METHOD
                    )
                req = Event()
                req.read(iprot)
                iprot.readMessageEnd()
            except EOFError:
                print "swallowing EOF"
            return event_handler(context, req)
        return event_created_callback
