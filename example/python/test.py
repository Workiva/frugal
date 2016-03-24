from datetime import timedelta
import json
import logging

from tornado import gen
from tornado import ioloop
from tornado import concurrent

from nats.io.client import Client as NATS
from nats.io.utils import new_inbox

# from frugal.context import FContext
logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)


class Heartbeat(object):

    def __init__(self, listen_subject, reply_subject, interval, missed_count=0):
        self.listen_subject = listen_subject
        self.reply_subject = reply_subject
        self.interval = interval
        self.missed_count = missed_count

    def increment_missed(self):
        self.missed_count += 1

    def reset_count(self):
        self.missed_count = 0


@gen.coroutine
def main():
    nats_client = NATS()
    options = {"verbose": True, "servers": ["nats://127.0.0.1:4222"]}

    yield nats_client.connect(**options)

    payload = json.dumps({"version": 0})

    inbox = "frugal." + new_inbox()

    future = concurrent.Future()
    sid = yield nats_client.subscribe(inbox, b'', None, future)
    yield nats_client.auto_unsubscribe(sid, 1)
    yield nats_client.publish_request("foo", inbox, payload)

    # Handshake call
    msg = yield gen.with_timeout(timedelta(milliseconds=10000), future)

    subjects = msg.data.split()

    hb = Heartbeat(subjects[0], subjects[1], int(subjects[2]))

    listen_to = msg.subject
    # write_to = msg.reply

    def on_message_cb(msg=None):
        if msg.reply == "DISCONNECT":
            print("Got disconnect message.")
            return
        # TODO write msg.data to writer

    yield nats_client.subscribe(listen_to, "", on_message_cb)

    # Setup Heartbeat
    def on_heartbeat_cb(msg=None):
        # Received a heartbeat, set missed count to 0
        print("received heartbeat. count: {}".format(hb.missed_count))
        hb.reset_count()

        if not heartbeat_timer.is_running():
            heartbeat_timer.start()

    @gen.coroutine
    def send_heartbeat(future=None):
        if hb.missed_count > 4:
            print("heartbeat expired...shut it down")
        hb.increment_missed()
        yield nats_client.publish(hb.reply_subject, "")
        if future is None:
            future = concurrent.Future()

    heartbeat_timer = ioloop.PeriodicCallback(send_heartbeat, hb.interval)

    # Subscribe to heartbeat listen
    yield nats_client.subscribe(hb.listen_subject, "", on_heartbeat_cb)

    # Start things off publishing
    # yield nats_client.publish(hb.listen_subject, "")


if __name__ == '__main__':
    io_loop = ioloop.IOLoop.instance()
    io_loop.add_callback(main)
    io_loop.start()
