from event.f_Foo import Iface


class ExampleHandler(Iface):

    def ping(self, ctx):
        print "ping: {}".format(ctx)

    def oneWay(self, ctx, id, req):
        print "oneWay: {} {} {}".format(ctx, id, req)

    def blah(self, ctx, num, Str, event):
        print "blah: {} {} {} {}".format(ctx, num, Str, event)
        ctx.set_response_header("foo", "bar")
        return 42

    def basePing(self, ctx):
        print "basePing: {}".format(ctx)

