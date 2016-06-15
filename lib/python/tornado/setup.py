from setuptools import setup, find_packages

from frugal.version import __version__

setup(
    name='frugal-tornado',
    version=__version__,
    description='Frugal Python Library',
    maintainer='Messaging Team',
    maintainer_email='messaging@workiva.com',
    url='http://github.com/Workiva/frugal',
    packages=find_packages(),
)
