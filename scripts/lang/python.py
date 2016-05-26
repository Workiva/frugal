import os
from shutil import copyfile

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

    def update_expected_tests(self, root):
        files_to_update = ['f_Blah.py',
                           'f_blah_publisher.py',
                           'f_blah_subscriber.py',
                           'f_Foo_publisher.py',
                           'f_Foo_subscriber.py']

        out = os.path.join(root, "test/out/valid")
        expected = os.path.join(root, "test/expected/python")

        for f in files_to_update:
            src = os.path.join(out, f)
            dest = os.path.join(expected, f)
            print "copying {} to {}".format(src, dest)
            copyfile(src, dest)
