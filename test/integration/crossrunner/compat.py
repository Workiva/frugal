#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
#

import os
import sys

if sys.version_info[0] == 2:
    _ENCODE = sys.getfilesystemencoding()

    def path_join(*args):
        bin_args = map(lambda a: a.decode(_ENCODE), args)
        return os.path.join(*bin_args).encode(_ENCODE)

    def str_join(s, l):
        bin_args = map(lambda a: a.decode(_ENCODE), l)
        b = s.decode(_ENCODE)
        return b.join(bin_args).encode(_ENCODE)

    logfile_open = open

else:

    path_join = os.path.join
    str_join = str.join

    def logfile_open(*args):
        return open(*args, errors='replace')