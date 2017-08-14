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

import webapp2

from frugal.server.http_handler import _FHttpRequest
from frugal.server.http_handler import _FSynchronousHttpRequestHandler


def new_webapp2_handler(processor, protocol_factory):
    """
    Produces a class extending webapp2.RequestHandler that can be used to
    handle frugal HTTP rpc requests.

    Args:
        processor: The processor to use to handle requests.
        protocol_factory: A protocol factory to serialize/deserialize frugal
                          requests.
    """
    handler = _FSynchronousHttpRequestHandler(processor, protocol_factory)

    class FWebapp2Handler(webapp2.RequestHandler):
        """
        FWebapp2Handler uses the webapp2 framework to handle frugal HTTP rpc
        requests.
        """
        def post(self):
            request = _FHttpRequest(
                headers=self.request.headers,
                body=self.request.body,
            )
            response = handler.handle_http_request(request)
            self.response.set_status(response.status_code)
            self.response.headers.update(response.headers)
            self.response.write(response.body)

    return FWebapp2Handler
