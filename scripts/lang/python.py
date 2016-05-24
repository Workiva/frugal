import os

from lang.base import LanguageBase


class Python(LanguageBase):
    """
    Python implementation of LanguageBase.
    """

    def update_frugal(self, version, root):
        """Update the Python version."""

        os.chdir('{0}/lib/python'.format(root))

        with open('frugal/version.py', 'w') as f:
            f.write("__version__ = '{0}'".format(version))

