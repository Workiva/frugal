class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request_send(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def __request_recv(self, ctx, method, result):
        self._iprot.read_response_headers(ctx)
        _, mtype, _ = self._iprot.readMessageBegin()
        if mtype == TMessageType.EXCEPTION:
            x = TApplicationException()
            x.read(self._iprot)
            self._iprot.readMessageEnd()
            if x.type == TApplicationExceptionType.RESPONSE_TOO_LARGE:
                raise TTransportException(type=TTransportExceptionType.RESPONSE_TOO_LARGE, message=x.message)
            raise x
        
        result.read(self._iprot)
        self._iprot.readMessageEnd()class FClient(object):
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def __request_recv(self,)class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def __class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __reclass FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()class FClient(object):
    def __init__(self, provider):
        # self._transport = provider.get_transport()
        # self._protocol_factory = provider.get_protocol_factory()
        # self._oprot = self._protocol_factory.get_protocol(self._transport)
        # self._iprot = self._protocol_factory.get_protocol(self._transport)
        # self._write_lock = Lock()class FClient(object):
    def __init__(self, provider):
        class FClient(object):
    def __init__(self, provide)class FClient(object):
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def __request_recv(self,)class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def __class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()

    def _class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        oprot = self._oprot
        with self._write_lock:
            oprot.get_transport().set_timeout2(ctx.timeout)
            oprot.write_request_headers(ctx)
            oprot.writeMessageBegin(method, TMessageType.CALL, 0)
            args.write(oprot)
            oprot.writeMessageEnd()
            oprot.get_transport().flush()class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __request(self, ctx, method, args):
        class FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()

    def __reclass FClient(object):
    def __init__(self, provider):
        self._transport = provider.get_transport()
        self._protocol_factory = provider.get_protocol_factory()
        self._oprot = self._protocol_factory.get_protocol(self._transport)
        self._iprot = self._protocol_factory.get_protocol(self._transport)
        self._write_lock = Lock()class FClient(object):
    def __init__(self, provider):
        # self._transport = provider.get_transport()
        # self._protocol_factory = provider.get_protocol_factory()
        # self._oprot = self._protocol_factory.get_protocol(self._transport)
        # self._iprot = self._protocol_factory.get_protocol(self._transport)
        # self._write_lock = Lock()class FClient(object):
    def __init__(self, provider):
        class FClient(object):
    def __init__(self, provide)