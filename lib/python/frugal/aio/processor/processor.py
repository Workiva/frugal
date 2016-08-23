import logging
from asyncio import Lock

from thrift.Thrift import TApplicationException, TMessageType, TType

from frugal.aio.processor import FProcessorFunction

logger = logging.getLogger(__name__)


class FProcessor(object):
    """FProcessor is a generic object which operates upon an input stream and
    writes to some output stream.
    """

    async def process(self, iprot, oprot):
        pass


class FBaseProcessor(FProcessor):

    def __init__(self):
        """Create new instance of FBaseProcessor that will process requests."""
        self._processor_function_map = {}
        self._write_lock = Lock()

    def add_to_processor_map(self, key: str, proc: FProcessorFunction):
        """Register the given FProcessorFunction.

        Args:
            key: processor function name
            proc: FProcessorFunction
        """
        self._processor_function_map[key] = proc

    def get_write_lock(self):
        """Return the write lock."""
        return self._write_lock

    async def process(self, iprot, oprot):
        """Process an input protocol and output protocol

        Args:
            iprot: input FProtocol
            oport: ouput FProtocol

        Raises:
            TApplicationException: if the processor does not know how to handle
                                   this type of function.
        """
        context = iprot.read_request_headers()
        name, _, _ = iprot.readMessageBegin()

        processor_function = self._processor_function_map.get(name)

        # If the function was in our dict, call process on it.
        if processor_function:
            try:
                return await processor_function.process(context, iprot, oprot)
            except Exception as e:
                logging.warn('frugal: error processing request with ' +
                             'correlation id %s: %s' %
                             (context.get_correlation_id(), e))
                raise

        iprot.skip(TType.STRUCT)
        iprot.readMessageEnd()

        ex = TApplicationException(TApplicationException.UNKNOWN_METHOD,
                                   "Unknown function: {0}".format(name))

        async with self._write_lock:
            oprot.write_response_headers(context)
            oprot.writeMessageBegin(name, TMessageType.EXCEPTION, 0)
            ex.write(oprot)
            oprot.writeMessageEnd()
            oprot.trans.flush()

        logger.exception(ex)
        raise ex
