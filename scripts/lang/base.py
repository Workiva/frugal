class LanguageBase(object):
    """Language update implementations must implement LanguageBase."""

    def update_frugal(self, version, root):
        """Update the frugal version."""
        raise Exception('update_frugal not implemented')
