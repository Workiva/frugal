"""
Use this tool to parse serialized Frugal protocol frames into their components.
"""

import argparse
import array
from struct import unpack_from
import sys


BYTES_INPUT = 'bytes'
STRING_INPUT = 'string'

V0_PROTOCOL = 0x00


parser = argparse.ArgumentParser(description='Parse Frugal frames')
parser.add_argument(
    '-i', '--input', type=str, default=BYTES_INPUT,
    choices=[BYTES_INPUT, STRING_INPUT],
    help='Determines the format of the input (default is byte list)')
parser.add_argument('-p', '--payload', action='store_true',
                    help='Print payload')
parser.add_argument('frame', type=str, help='Frugal frame')


def parse_bytes(frame, print_payload=False):
    """Parse the frame as a list of bytes, e.g. '[0 0 0 70 0 0 0 0 48]'."""

    frame = frame.strip()
    if frame.startswith('['):
        frame = frame[1:]
    if frame.endswith(']'):
        frame = frame[:-1]
    frame = frame.replace(',', ' ')
    frame = array.array('B', [int(b) for b in frame.split()]).tostring()
    parse_string(frame, print_payload=print_payload)


def parse_string(frame, print_payload=False):
    """Parse the frame as string."""

    # Need at least 5 bytes (4 for frame size, 1 for protocol version)
    if len(frame) < 5:
        print 'Invalid frame size {}'.format(len(frame))
        sys.exit(1)

    frame_size = unpack_from('!I', frame[:4])[0]
    if len(frame[4:]) != frame_size:
        print 'Frame size {} does not match actual frame length {}'.format(
            frame_size, len(frame[4:]))
        sys.exit(1)

    version = unpack_from('!B', frame[0:1])[0]
    headers = None
    payload = None

    if version == V0_PROTOCOL:
        headers, payload = parse_v0_protocol(frame[5:])
    else:
        print 'Invalid protocol version {}'.format(version)
        sys.exit(1)

    print_frame(frame_size, version, headers, payload,
                print_payload=print_payload)


def print_frame(frame_size, version, headers, payload, print_payload=False):
    payload = array.array('B', payload)
    print 'Frame size:\t\t{}'.format(frame_size)
    print 'Protocol version:\t{}'.format(version)
    print 'Headers:\t\t{}'.format(headers)
    print 'Payload size:\t\t{}'.format(len(payload))
    if print_payload:
        print 'Payload:\t\t{}'.format(payload.tostring())
        print 'Payload bytes:\t\t{}'.format(payload.tolist())


def parse_v0_protocol(frame):
    """Parse the frame using the v0 protocol."""

    headers_size = unpack_from('!I', frame[0:4])[0]
    if len(frame[4:]) < headers_size:
        print 'Headers size {} less than remaining frame size {}'.format(
            headers_size, len(frame[4:]))
        sys.exit(1)

    headers = {}
    i = 4
    end = headers_size + 4
    while i < end:
        name_size = unpack_from('!I', frame[i:i + 4])[0]
        i += 4
        if i > end or i + name_size > end:
            print 'Invalid protocol header name size {}'.format(name_size)
            sys.exit(1)

        name = unpack_from('>{0}s'.format(name_size),
                           frame[i:i + name_size])[0]
        i += name_size
        if i > end:
            print 'Invalid protocol headers'
            sys.exit(1)

        val_size = unpack_from('!I', frame[i:i + 4])[0]
        i += 4
        if i > end or i + val_size > end:
            print 'Invalid protocol header value size {}'.format(val_size)
            sys.exit(1)

        value = unpack_from('>{0}s'.format(val_size), frame[i:i + val_size])[0]
        i += val_size
        if i > end:
            print 'Invalid protocol headers'
            sys.exit(1)

        headers[name] = value

    payload = frame[end:]
    return headers, payload


def main():
    args = parser.parse_args()
    frame = args.frame
    if args.input == BYTES_INPUT:
        parse_bytes(frame, print_payload=args.payload)
    elif args.input == STRING_INPUT:
        parse_string(frame, print_payload=args.payload)


if __name__ == '__main__':
    main()

