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
        "w-thrift==1.0.0-dev5",
        "requests==2.12.5",
    ],
    extras_require={
        'tornado': ["nats-client==0.5.0"],
        'asyncio': ["async-timeout==1.1.0", "asyncio-nats-client==0.3.1",
                    "aiohttp==0.22.3"],
        'gae': ["webapp2==2.5.2"],
    }
)
