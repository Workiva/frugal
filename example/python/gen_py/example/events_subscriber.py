from thrift.Thrift import TType, TApplicationException
from tornado import gen

from .ttypes import Event
from frugal.subscription import FSubscription


class EventsSubscriber(object):

    _DELIMETER = "."

    def __init__(self, provider):
        self._provider = provider

    @gen.coroutine
    def subscribe_event_created(self, user, event_handler):
        op = "EventCreated"
        prefix = "foo.{}.".format(user)
        topic = "{}Events{}{}".format(prefix, self._DELIMETER, op)

        transport, protocol = self._provider.new()

        yield transport.subscribe(topic, self.recv_EventCreated(protocol,
                                                                op,
                                                                event_handler))

    def recv_EventCreated(self, iprot, op, event_handler):
        def event_created_callback(msg=None):
            context = iprot.read_request_headers()
            (mname, mtype, mid) = iprot.readMessageBegin()
            if mname != op:
                iprot.skip(TType.STRUCT)
                iprot.readMessageEnd()
                raise TApplicationException(
                    TApplicationException.UNKNOWN_METHOD
                )
            req = Event()
            req.read(iprot)
            iprot.readMessageEnd()
            return event_handler(context, req)
        return event_created_callback
