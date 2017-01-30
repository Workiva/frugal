from thrift.transport.TTransport import TTransportBase
from thrift.transport.TTransport import TTransportException

from frugal.exceptions import FrugalTTransportExceptionType


class FTransport(object):
    """
    FTransport is comparable to Thrift's TTransport in that it represent the
    transport layer for frugal clients. However, frugal is callback based and
    sends only framed data. Therefore, instead of exposing read, write, and
    flush, the transport has a simple request method that sends framed frugal
    messages and returns the response.
    """
    def __init__(self, request_size_limit=0):
        """
        Args:
            request_size_limit: The maximum request payload size for this
                                transport. A non-positive number indicates
                                an unbounded allowable size.
        """
        self._request_size_limit = request_size_limit

    def open(self):
        """Open the transport."""
        raise NotImplementedError('You must override this')

    def close(self):
        """Close the transport."""
        raise NotImplementedError('You must override this')

    def is_open(self):
        """Return True if the transport is open, False otherwise."""
        raise NotImplementedError('You must override this')

    def set_monitor(self, monitor):
        """
        Set the transport monitor for the transport. This should only be used
        for "stateful" transports.

        Args:
            monitor: A transport monitor
        """
        raise NotImplementedError('You must override this')

    def oneway(self, context, payload):
        """
        Send the given framed frugal payload over the transport.

        Args:
            context:    FContext associated with the request (used for timeout
                        and logging)

            payload:    framed frugal data
        """
        raise NotImplementedError('You must override this')

    def request(self, context, payload):
        """
        Send the given framed frugal payload over the transport and returns the
        response.

        Args:
            context:    FContext associated with the request (used for timeout
                        and logging)

            payload:    framed frugal data
        """
        raise NotImplementedError('You must override this')

    def get_request_size_limit(self):
        """
        Returns the maximum request payload size for this transport. A non-
        positive number is returned to indicate an unbounded allowable size.
        """
        return self._request_size_limit

    def _preflight_request_check(self, payload):
        """
        Helper function that throws FrugalTTransportExceptionType.NOT_OPEN if the
        transport is not open or throws FMessageSizeException if the payload is
        too large. Should only be called by extending classes.
        """
        if not self.is_open():
            raise TTransportException(
                type=FrugalTTransportExceptionType.NOT_OPEN,
                message='Transport is not open')

        if len(payload) > self.get_request_size_limit() > 0:
            raise TTransportException(
                type=FrugalTTransportExceptionType.REQUEST_TOO_LARGE,
                message='Message exceeds {0} bytes, was {1} bytes'.format(
                    self.get_request_size_limit(), len(payload)))


class TSynchronousTransport(TTransportBase, object):
    """TSynchronousTransport is a Thrift TTransport for services which make
    synchronous requests.
    """

    def set_timeout(self, timeout):
        """Set the request timeout.

        Args:
            timeout: request timeout in milliseconds.
        """
        pass

