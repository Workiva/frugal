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


class FSubscription(object):
    """
    FSubscription to a pub/sub topic.  This is used only by generated code
    and should not be called directly.
    """

    def __init__(self, topic, transport):
        """
        Initialize FSubscription.

        Args:
            topic: pub/sub topic string.
            transport: FScopeTransport for the subscription.
        """
        self._topic = topic
        self._transport = transport

    def get_topic(self):
        """
        Return subscription topic.
        """
        return self._topic

    def unsubscribe(self):
        """
        Unsubscribe from the topic.

        The result of this is a future that should be awaited/yielded
        appropriately.
        """
        return self._transport.unsubscribe()

    def remove(self):
        """
        Unsubscribe and removes durably stored information on the broker,
        if applicable.
        """
        return self._transport.remove()
