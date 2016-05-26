import os
from xml.etree import ElementTree

from lang.base import LanguageBase


_POM = 'pom'
_POM_XML = 'pom.xml'
_WORKIVA = 'com.workiva'
_NS = {_POM: 'http://maven.apache.org/POM/4.0.0'}


class Java(LanguageBase):
    """
    Java implementation of LanguageBase. Uses xml.tree.ElementTree to parse all
    pom.xml's.
    """

    def update_frugal(self, version, root):
        """Update the java version."""
        # Update library pom
        os.chdir('{0}/lib/java'.format(root))
        self._update_maven_version(version)

        # Update example pom
        os.chdir('{0}/example/java'.format(root))
        self._update_maven_version(version)
        self._update_maven_dep(_WORKIVA, 'frugal', version)

    def _update_maven_version(self, version):
        """Update the project version in the current directory's pom.xml."""
        tree = ElementTree.parse(_POM_XML)
        ver = tree.getroot().find('{0}:version'.format(_POM), _NS)
        ver.text = version
        tree.write(_POM_XML, default_namespace=_NS[_POM])

    def _update_maven_dep(self, group, artifact, version):
        """Update a maven dependency in the current directory's pom.xml."""
        tree = ElementTree.parse(_POM_XML)
        for dep in tree.getroot().find('{0}:dependencies'.format(_POM), _NS):
            g = dep.find('{0}:groupId'.format(_POM), _NS)
            a = dep.find('{0}:artifactId'.format(_POM), _NS)
            if g.text == group and a.text == artifact:
                dep.find('{0}:version'.format(_POM), _NS).text = version
        tree.write(_POM_XML, default_namespace=_NS[_POM])

    def update_expected_tests(self, root):
        pass
