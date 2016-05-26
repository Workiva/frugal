
class FServer(object):
    """Base interface for a server, which must have a serve() method."""

    def serve(self):
        pass

    def stop(self):
        pass
