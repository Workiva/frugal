# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
        "six>=1.10.0,<2",
        "thrift==0.10.0",
        "requests>=2.12.5,<3",
    ],
    extras_require={
        'tornado': ["nats-client==0.7.2"],
        'asyncio': [
            "async-timeout==2.0.1",
            "asyncio-nats-client==0.7.2",
            # DO NOT CHANGE without verifying asyncio nats with tls works
            "multidict==3.1.3",
            "aiohttp==3.1.2",
        ],
        'gae': ["webapp2==2.5.2"],
    }
)
