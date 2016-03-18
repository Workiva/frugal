

class EventsPublisher(object):

    _DELIMETER = "."

    def __init__(self, scope_provider):
        """Creates an instance of EventsPublisher

        Args:
            scope_provider: FScopeProvider
        """
        self._scope_provider = scope_provider

    def open(self):
        trans, prot = self._scope_provider.new()
        self._trans = trans
        self._prot = prot
        self._trans.open()

    def close(self):
        self._trans.close()

    def publish_event_created(self, ctx, user, req):
        op = "EventCreated"

        self._trans.lock_topic(topic)
        # try or with, might need context manager
        self._prot.write_request_headers(ctx)
        self._prot.writeMessageBegin()
        req.write(self._prot)
        self._prot.writeMessageEnd()
        self._trans.flush()
        # finally
        self._trans.unlock_topic()
