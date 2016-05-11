from threading import Lock

from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from thrift.Thrift import TType


class FProcessor(object):
    """
    FProcessor is a generic object which operates upon an
    input stream and writes to some output stream.
    """

    def process(self, iprot, oprot):
        pass


class FBaseProcessor(FProcessor):

    def __init__(self, processor_function_map=None):
        """ Create new instance of FBaseProcessor that will process requests

        Args:
            processor_function_map: dict keyed by rpc call name for
                                    processor functions
        """
        if processor_function_map is None:
            processor_function_map = {}

        self._processor_function_map = processor_function_map
        self._write_lock = Lock()

    def process(self, iprot, oprot):
        """ Process an input protocol and output protocol

        Args:
            iprot: input FProtocol
            oport: ouput FProtocol

        Raises:
            TApplicationException: if the processor does not know how to handle
                this type of function.
        """

        context = iprot.read_request_header()
        (name, type, seqid) = iprot.readMessageBegin()

        processor_function = self._processor_function_map.get(name)

        # If the function was in our dict, call it
        if processor_function:
            processor_function(context, iprot, oprot)
            return

        iprot.skip(TType.STRUCT)
        iprot.readMessageEnd()

        ex = TApplicationException(TApplicationException.UNKNOWN_METHOD,
                                   "Unknown function: {0}".format(name))

        with self._write_lock:
            oprot.write_response_headers(context)
            oprot.writeMessageBegin(name, TMessageType.EXCEPTION, 0)

            ex.write(oprot)
            oprot.writeMessageEnd()
            oprot.trans.flush()

        raise ex
