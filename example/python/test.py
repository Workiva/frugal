from datetime import timedelta
import json

from tornado import gen
from tornado import ioloop
from tornado import concurrent

from nats.io.client import Client as NATS
from nats.io.utils import new_inbox

# from frugal.context import FContext


@gen.coroutine
def main():

    # context = FContext()

    nats_client = NATS()
    options = {"verbose": True, "servers": ["nats://127.0.0.1:4222"]}

    yield nats_client.connect(**options)

    payload = json.dumps({"version": 0})

    inbox = "frugal." + new_inbox()

    future = concurrent.Future()
    sid = yield nats_client.subscribe(inbox, b'', None, future)
    yield nats_client.auto_unsubscribe(sid, 1)
    yield nats_client.publish_request("foo", inbox, payload)
    msg = yield gen.with_timeout(timedelta(milliseconds=5000), future)
    print("message: {0}".format(dir(msg)))
    print("message data: {0}".format(msg.data))

    subjects = msg.data.split()
    heartbeat_listen = subjects[0]
    heartbeat_reply = subjects[1]
    heartbeat_interval = int(subjects[2])

    print("heartbeat listen {0}".format(heartbeat_listen))
    print("heartbeat reply {0}".format(heartbeat_reply))
    print("heartbeat interval {0}".format(heartbeat_interval))

    print("listenTo: {0}".format(msg.subject))
    print("writeTo: {0}".format(msg.reply))


if __name__ == '__main__':
    ioloop.IOLoop.instance().run_sync(main)

