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
        "requests==2.12.5",
    ],
    extras_require={
        'tornado': ["nats-client==0.3.0"],
        'asyncio': ["async-timeout==1.1.0", "asyncio-nats-client==0.3.1",
                    "aiohttp==0.22.3"],
        'gae': ["webapp2==2.5.2"],
    }
)
