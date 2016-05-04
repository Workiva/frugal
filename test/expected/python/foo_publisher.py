#
# Autogenerated by Frugal Compiler (1.3.0)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



from thrift.Thrift import TMessageType




class FooPublisher(object):
    """
    And this is a scope docstring.
    """

    _DELIMITER = '.'

    def __init__(self, provider):
        """
        Create a new FooPublisher.

        Args:
            provider: FScopeProvider
        """

        self._transport, self._protocol = provider.new()

    def open(self):
        self._transport.open()

    def close(self):
        self._transport.close()

    def publish_Foo(self, ctx, baz, req):
        """
        This is an operation docstring.
        
        Args:
            ctx: FContext
            baz: string
            req: Thing
        """

        op = 'Foo'
        prefix = 'foo.bar.%s.qux.' % (baz)
        topic = '%sFoo%s%s' % (prefix, self._DELIMITER, op)
        oprot = self._protocol
        self._transport.lock_topic(topic)
        try:
            oprot.writeRequestHeader(ctx)
            oprot.writeMessageBegin(op, TMessageType.CALL, 0)
            req.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()
        finally:
            self._transport.unlock_topic()


    def publish_Bar(self, ctx, baz, req):
        """
        Args:
            ctx: FContext
            baz: string
            req: Stuff
        """

        op = 'Bar'
        prefix = 'foo.bar.%s.qux.' % (baz)
        topic = '%sFoo%s%s' % (prefix, self._DELIMITER, op)
        oprot = self._protocol
        self._transport.lock_topic(topic)
        try:
            oprot.writeRequestHeader(ctx)
            oprot.writeMessageBegin(op, TMessageType.CALL, 0)
            req.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()
        finally:
            self._transport.unlock_topic()

