#!/usr/bin/env python
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

# Frugal - integration test suite
#
# tests various server-client, protocol and transport combinations
#
# This script supports python 2.7 and later.
# python 3.x is recommended for better stability.
#

from __future__ import print_function
from itertools import chain
import json
import logging
import multiprocessing
import argparse
import os
import sys

import crossrunner
from crossrunner.compat import path_join

ROOT_DIR = os.path.dirname(os.path.realpath(os.path.dirname(__file__)))
TEST_DIR_RELATIVE = 'integration'
TEST_DIR = path_join(ROOT_DIR, TEST_DIR_RELATIVE)
DEFAULT_CONFIG_FILE = 'tests.json'

def run_cross_tests(server_match, client_match, jobs, retry_count, regex, config_file):
    logger = multiprocessing.get_logger()
    logger.debug('Collecting tests')

    with open(path_join(TEST_DIR, config_file), 'r') as fp:
        j = json.load(fp)
    tests = crossrunner.collect_cross_tests(j, server_match, client_match, regex)

    if not tests:
        print('No test found that matches the criteria', file=sys.stderr)
        print('  servers: %s' % server_match, file=sys.stderr)
        print('  clients: %s' % client_match, file=sys.stderr)
        return False

    dispatcher = crossrunner.TestDispatcher(TEST_DIR, ROOT_DIR, TEST_DIR_RELATIVE, jobs)
    logger.debug('Executing %d tests' % len(tests))
    try:
        for r in [dispatcher.dispatch(test, retry_count) for test in tests]:
            r.wait()
        logger.debug('Waiting for completion')
        return dispatcher.wait()
    except (KeyboardInterrupt, SystemExit):
        logger.debug('Interrupted, shutting down')
        dispatcher.terminate()
        return False


def main(argv):
    parser = argparse.ArgumentParser()
    parser.add_argument('--server', default='', nargs='*',
                        help='list of servers to test')
    parser.add_argument('--client', default='', nargs='*',
                        help='list of clients to test')
    parser.add_argument('-R', '--regex', help='test name pattern to run')

    g = parser.add_argument_group(title='Advanced')
    g.add_argument('-v', '--verbose', action='store_const',
                   dest='log_level', const=logging.DEBUG, default=logging.WARNING,
                   help='show debug output for test runner')
    options = parser.parse_args(argv)

    logger = multiprocessing.log_to_stderr()
    logger.setLevel(options.log_level)

    server_match = list(chain(*[x.split(',') for x in options.server]))
    client_match = list(chain(*[x.split(',') for x in options.client]))

    options.jobs = int(multiprocessing.cpu_count()) - 1
    options.retry_count = 0

    config_file = DEFAULT_CONFIG_FILE

    res = run_cross_tests(server_match, client_match, options.jobs, options.retry_count, options.regex, config_file)
    return 0 if res else 1


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
