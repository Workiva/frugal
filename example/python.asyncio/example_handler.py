from event.f_Foo import Iface


class ExampleHandler(Iface):

    async def ping(self, ctx):
        print("ping: {}".format(ctx))

    async def oneWay(self, ctx, id, req):
        print("oneWay: {} {} {}".format(ctx, id, req))

    async def blah(self, ctx, num, Str, event):
        print("blah: {} {} {} {}".format(ctx, num, Str, event))
        ctx.set_response_header("foo", "bar")
        return 42

    async def basePing(self, ctx):
        print("basePing: {}".format(ctx))

