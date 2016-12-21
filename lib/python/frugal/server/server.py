import logging
import struct
from frugal import _NATS_MAX_MESSAGE_SIZE

logger = logging.getLogger(__name__)

class FServer(object):
    """Base interface for a server, which must have a serve() method."""

    def serve(self):
        pass

    def stop(self):
        pass


class FNatsBaseServer(FServer):
    """Base interface for a nats server, with some utility methods."""


    def _validate_nats_message(self, msg):
        """
        Verifies that the received message has a replyto and a valid framesize.

        :param msg: Nats message received by the
        :return: True if message is valid, False if not
        """

        reply_to = msg.reply
        if not reply_to:
            logger.warn("Discarding invalid NATS request (no reply)")
            return False

        frame_size = struct.unpack('!I', msg.data[:4])[0]
        if frame_size > _NATS_MAX_MESSAGE_SIZE - 4:
            logger.warning("Invalid frame size, dropping message.")
            return False

        return True

    def _is_message_oneway(self, otrans):
        """
        Checks whether a received message is a one-way message.

        :param otrans: output transport
        :return: True if message response indicates 1-way, False if not
        """

        # A frame with length 4 indicates a 1-way message as it only includes
        # the frame size with no data
        if len(otrans) == 4:
            return True
        return False
