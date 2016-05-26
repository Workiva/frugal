from setuptools import setup, find_packages

from frugal.version import __version__

setup(
    name='frugal',
    version=__version__,
    description='Frugal Python Library',
    maintainer='Charlie Strawn',
    maintainer_email='charlie.strawn@workiva.com',
    url='http://github.com/Workiva/frugal',
    packages=find_packages(),
)
