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


class FTransportFactory(object):
    """
    FTransportFactory is responsible for creating new FTransports.
    """

    def get_transport(self, thrift_transport):
        """
        Retuns a new FTransport wrapping the given TTransport.

        Args:
            thrift_transport: TTransport to wrap.
        Returns:
            new FTranpsort
        """
        pass


class FPublisherTransportFactory(object):
    """
    FPublisherTransportFactory is responsible for creating new
    FPublisherTransports.
    """

    def get_transport(self):
        """
        Returns a new FPublisherTransport.
        """
        pass


class FSubscriberTransportFactory(object):
    """
    FSubscriberTransportFactory is responsible for creating new
    FSubscriberTransports.
    """

    def get_transport(self):
        """
        Returns a new FSubscriberTransport.
        """
        pass
