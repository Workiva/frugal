import os
from shutil import copyfile, move
from tempfile import mkstemp

from lang.base import LanguageBase


class Python(LanguageBase):
    """
    Python implementation of LanguageBase.
    """

    def update_frugal(self, version, root):
        """Update the Python version."""

        print "Python: Updating frugal core to version: {}.".format(version)
        os.chdir('{0}/lib/python/core'.format(root))
        with open('frugal/version.py', 'w') as f:
            f.write("__version__ = '{0}'".format(version))

        print "Python: Updating frugal tornado to version: {}.".format(version)
        os.chdir('{0}/lib/python/tornado'.format(root))
        with open('frugal_tornado/version.py', 'w') as f:
            f.write("__version__ = '{0}'".format(version))
        self._update_tornado_requirements(version, root)

    def _update_tornado_requirements(self, version, root):
        os.chdir('{0}/lib/python/tornado/'.format(root))
        requirements_file = os.path.join(os.getcwd(), 'requirements.txt')
        self._replace(requirements_file, "frugal==", version)

    def update_expected_tests(self, root):
        files_to_update = ['f_Blah.py',
                           'f_blah_publisher.py',
                           'f_blah_subscriber.py',
                           'f_Foo_publisher.py',
                           'f_Foo_subscriber.py']

        valid = os.path.join(root, "test/out/valid")
        expected = os.path.join(root, "test/expected/python")

        for f in files_to_update:
            src = os.path.join(expected, f)
            dest = os.path.join(valid, f)
            print "copying {} to {}".format(src, dest)
            copyfile(src, dest)

    def _replace(self, file_path, pattern, version):
        fh, abs_path = mkstemp()
        with open(abs_path, 'w') as new_file:
            with open(file_path) as old_file:
                for line in old_file:
                    if pattern in line:
                        new_file.write("{}{}\n".format(pattern, version))
                    else:
                        new_file.write(line)
        os.close(fh)
        os.remove(file_path)
        move(abs_path, file_path)
