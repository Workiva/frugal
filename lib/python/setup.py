from setuptools import setup, find_packages

from frugal import VERSION

setup(
    name='frugal',
    version=VERSION,
    description='Frugal Python Library',
    maintainer='Charlie Strawn',
    maintainer_email='charlie.strawn@workiva.com',
    url='http://github.com/Workiva/frugal',
    packages=find_packages(),
)
