from threading import Thread, Lock

from thrift.transport.TTransport import TFramedTransport

from .transport import FTransport


class FMuxTransport(FTransport):

    def __init__(self, thrift_transport, num_workers):
        """Construct a new FMuxTransport

        Args:
            thrift_transport: TTransport to wrap
            num_workers: number of worker threads for the FTranpsort
        """
        self._framed_transport = TFramedTransport(thrift_transport)
        self._work_queue = ArrayBlockingQueue(num_workers)
        self._processor_thread = ProcessorThread()
        self._worker_threads = WorkerThread(num_workers)
        self._lock = Lock()

    def set_registry(self, registry):
        with self._lock:
            if registry is None:
                raise StandardError("registry cannot be null.")

            if self._registry is not None:
                return

            self._registry = registry

    def is_open(self):
        with self._lock:
            trans_open = self._framed_transport.isOpen()
            return trans_open and self._registery is not None

    def open(self):
        with self._lock:
            self._framed_transport.open()

    def close(self):
        with self._lock:
            if self._registry is None:
                return

            self._framed_transport.close()

    def read(self, size):
        self._framed_transport.read(size)

    def write(self, buff):
        self._framed_transport.write(buff)

    def flush(self):
        with self._lock:
            self._framed_transport.flush()


class ProcessorThread(Thread):

    def __init__(self, trans):
        Thread.__init__()
        self._running = False
        self._trans = trans

    def run(self):
        """Not much documentation for this.  Going to have to experiment."""

        self._running = True
        while self._running:
            self._trans.readFrame()


