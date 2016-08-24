from setuptools import setup, find_packages

from frugal.version import __version__

setup(
    name='frugal',
    version=__version__,
    description='Frugal Python Library',
    maintainer='Messaging Team',
    maintainer_email='messaging@workiva.com',
    url='http://github.com/Workiva/frugal',
    packages=find_packages(exclude=('frugal.tests', 'frugal.tests.*')),
    install_requires=[
        "w-thrift==1.0.0-dev5",
    ],
    extras_require={
        'tornado': ["nats-client==0.2.4"],
        'asyncio': ["asyncio-nats-client==0.3.1", "aiohttp==0.22.3"]
    }
)
