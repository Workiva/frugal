import os

from yaml import dump, load

from lang.base import LanguageBase


class Dart(LanguageBase):
    """
    Dart implementation of LanguageBase. Uses PyYAML to parse all
    pubspec.yaml's.
    """

    def update_frugal(self, version, root):
        """Update the dart version."""
        # Update libary pubspec
        def update_lib(data):
            data['version'] = version
        os.chdir('{0}/lib/dart'.format(root))
        self._update(update_lib, 'Dart lib')

        # Update example pubspec
        def update_example(data):
            data['dependencies']['frugal']['version'] = '^{0}'.format(
                version
            )
        os.chdir('{0}/examples/dart'.format(root))
        self._update(update_example, 'Dart example')

    def _update(self, update, prefix):
        """
        Update pubspec.yaml in current directory using the given update
        function.
        """
        pubspec = 'pubspec.yaml'
        with open(pubspec, 'r') as f:
            data = load(f.read())
            update(data)
        with open(pubspec, 'w') as f:
            dump(data, f, default_flow_style=False)

    def update_expected_tests(self, root):
        pass
