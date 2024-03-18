#! /usr/bin/env python

import argparse
import re
import subprocess
from xml.etree import ElementTree

from yaml import dump, safe_load


_PYTHON_VERSION_FILE = "lib/python/frugal/version.py"
_POM = "pom"
_POM_XML = "{0}/pom.xml"
_NS = {_POM: "http://maven.apache.org/POM/4.0.0"}


class LanguageBase(object):
    """Language update implementations must implement LanguageBase."""

    def update_frugal(self, version):
        """Update the frugal version."""
        raise Exception("update_frugal not implemented")


class Python(LanguageBase):
    """
    Python implementation of LanguageBase.
    """

    def update_frugal(self, version):
        """Update the Python version."""

        old_stuff = ""
        with open(_PYTHON_VERSION_FILE, "r") as f:
            for line in f:
                if "__version__" in line:
                    old_stuff += "__version__ = '{0}'\n".format(version)
                    break
                old_stuff += line

        with open(_PYTHON_VERSION_FILE, "w") as f:
            f.write(old_stuff)


class Dart(LanguageBase):
    """
    Dart implementation of LanguageBase. Uses PyYAML to parse all
    pubspec.yaml's.
    """

    def update_frugal(self, version):
        """Update the dart version."""
        # Update libary pubspec
        def update_lib(data):
            data["version"] = version

        self._update("lib/dart", update_lib, "Dart lib")

        # Update example pubspec
        def update_example(data):
            data["dependencies"]["frugal"]["version"] = "^{0}".format(version)

        self._update("examples/dart", update_example, "Dart example")

    def _update(self, where, update, prefix):
        """
        Update pubspec.yaml in current directory using the given update
        function.
        """
        pubspec = where + "/pubspec.yaml"
        with open(pubspec, "r") as f:
            data = safe_load(f.read())
            update(data)
        with open(pubspec, "w") as f:
            dump(data, f, default_flow_style=False)


class Java(LanguageBase):
    """
    Java implementation of LanguageBase. Uses xml.tree.ElementTree to parse all
    pom.xml's.
    """

    def update_frugal(self, version):
        """Update the java version."""
        # Update library pom
        self._update_maven_version("lib/java", version)

        # Update example pom
        self._update_maven_version("examples/java", version)
        self._update_maven_dep("examples/java", version)

        # Update integration tests
        integration = "test/integration/java/frugal-integration-test"
        self._update_maven_dep(integration, version)

    def _update_maven_version(self, where, version):
        """Update the project version in the current directory's pom.xml."""
        pwd = _POM_XML.format(where)
        tree = ElementTree.parse(pwd)
        ver = tree.getroot().find("{0}:version".format(_POM), _NS)
        ver.text = version
        tree.write(pwd, default_namespace=_NS[_POM])

    def _update_maven_dep(self, where, version):
        """Update a maven dependency in the current directory's pom.xml."""
        pwd = _POM_XML.format(where)
        tree = ElementTree.parse(pwd)
        for dep in tree.getroot().find("{0}:dependencies".format(_POM), _NS):
            g = dep.find("{0}:groupId".format(_POM), _NS)
            a = dep.find("{0}:artifactId".format(_POM), _NS)
            if g.text == "com.workiva" and a.text == "frugal":
                dep.find("{0}:version".format(_POM), _NS).text = version
        tree.write(pwd, default_namespace=_NS[_POM])


class Go(LanguageBase):
    """
    Go implementation of LanguageBase.
    """

    def update_frugal(self, version):
        # No change necessary
        pass


LANGUAGES = {
    "dart": Dart(),
    "go": Go(),
    "java": Java(),
    "python": Python(),
}

_VERSION_MATCH = ".*?..*?..*?"


def main(args):
    update_frugal_version(args.version.strip("v"))


def update_frugal_version(version):
    """Update the frugal version."""
    # TODO: Implement dry run
    print(f"Updating frugal to {version} for {', '.join(LANGUAGES.keys())}")
    update_compiler(version)
    install_frugal()
    for lang in LANGUAGES.values():
        lang.update_frugal(version)
    update_tests()
    update_examples()


def update_compiler(version):
    """Update the frugal compiler."""
    # Update the global version
    base_str = 'const Version = "{0}"'
    sub_str = base_str.format(_VERSION_MATCH)
    ver_str = base_str.format(version)
    glob = "compiler/globals/globals.go"
    s = ""
    with open(glob, "r") as f:
        s = re.sub(sub_str, ver_str, f.read())
    with open(glob, "w") as f:
        f.write(s)


def install_frugal():
    # Install the binary with the updated version
    if subprocess.call(["go", "mod", "download"]) != 0:
        raise Exception("downloading modules failed")
    if subprocess.call(["go", "install"]) != 0:
        raise Exception("installing frugal binary failed")


def update_tests():
    """Update the frugal generation tests."""
    if subprocess.call(["go", "test", "./compiler", "--copy-files"]) != 0:
        raise Exception("Failed to update generated tests")
    if subprocess.call(
        [
            "frugal",
            "--gen",
            "dart:use_enums=true",
            "-r",
            "--out=dart/gen-dart",
            "frugalTest.frugal",
        ],
        cwd="./test/integration",
    ):
        raise Exception("Failed to generate Dart test code")


def update_examples():
    """Update the examples."""
    # TODO: Replace the generate example shell script
    if subprocess.call(
        ["make", "generate"],
        stdout=subprocess.DEVNULL,
        cwd="./examples",
    ):
        raise Exception("Failed to generate example code")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Update version")
    parser.add_argument("--version", dest="version", type=str, required=True)
    args = parser.parse_args()
    main(args)
