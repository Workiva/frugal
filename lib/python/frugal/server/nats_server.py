import logging
from . import FServer

logger = logging.getLogger(__name__)

# TODO: implement this.


class FNatsServer(FServer):

    def __init__(self):
        logger.exception()

    def serve(self):
        pass
        # self.connection.QueueSubscribe(self.subject, queue, callback)

