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

        yield transport.subscribe(topic)

    def recv_EventCreated(op, iprot):
        name, type, seqid = iprot.readMessageBegin()
        if name != op:
            iprot.skip(TType.STRUCT)
            iprot.readMessageEnd()
            raise TApplicationException(TApplicationException.UNKNOWN_METHOD)
        req = Event()
        req.read(iprot)
        iprot.readMessageEnd()
        return req
