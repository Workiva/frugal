from tornado import gen


class FProcessorFunction(object):

    @gen.coroutine
    def process(self, ctx, iprot, oprot):
        pass
