import os

from lang.base import LanguageBase


class Python(LanguageBase):
    """
    Python implementation of LanguageBase.
    """

    def update_frugal(self, version, root):
        """Update the Python version."""

        os.chdir('{0}/lib/python'.format(root))

        old_stuff = ""
        with open('frugal/version.py', 'r') as f:
            for line in f:
                if "__version__" in line:
                    old_stuff += "__version__ = '{0}'\n".format(version)
                    break
                old_stuff += line

        with open('frugal/version.py', 'w') as f:
            f.write(old_stuff)

    def update_expected_tests(self, root):
        pass
