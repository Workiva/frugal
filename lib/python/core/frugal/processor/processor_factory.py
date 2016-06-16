

class FProcessorFactory(object):
    """FProcessorFactory creates FProcessors.  The default factory just returns
    a singleton.
    """

    def __init__(self, processor):
        """Initialize factory and set processor instance.

        Args:
            processor: FProcessor singleton instance.
        """

        self._processor = processor

    def get_processor(self, transport):
        """Return processor.

        Args:
            transport: TTransport
        """

        return self._processor

