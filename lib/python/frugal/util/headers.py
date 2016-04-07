from io import BytesIO
from struct import pack_into, unpack_from

from frugal.exceptions import FrugalVersionException


_V0 = 0
_UCHAR = '>B'
_UINT = '>I'


class Writer(object):

    def __init__(self):
        self.buff = BytesIO()

    def write_headers_to_buffer(self, headers):
        """Writes a given dictionary to a bytearray object and returns it

        Args:
            headers: dict of frugal headers to write
        Returns:
            bytearray containing binary headers
        """
        size = self._compute_size(headers)
        offset = 5

        buff = bytearray(size + offset)

        pack_into(_UCHAR, buff, 0, _V0)
        pack_into(_UINT, buff, 1, size)

        for key, value in headers.iteritems():
            pack_into(_UINT, buff, offset, len(key))
            offset += 4

            pack_into('>{0}s'.format(str(len(key))), buff, offset, key)
            offset += len(key)

            pack_into(_UINT, buff, offset, len(value))
            offset += 4

            pack_into('>{0}s'.format(str(len(value))), buff, offset, value)
            offset += len(value)

        return buff

    def write_bytes(self, buff):
        self.buff.write(buff)

    def get_buffer(self):
        return self.buff.getvalue()

    def _compute_size(self, headers):
        size = 0
        for key, value in headers.iteritems():
            size = size + 8 + len(key) + len(value)
        return size


class Reader(object):

    def __init__(self, writer=None):
        self.writer = writer or Writer()

    def read_request_headers(self):
        return self._read_headers(self.writer.get_buffer())

    def _read_headers(self, buff):
        parsed_headers = {}

        version = unpack_from(_UCHAR, buff, 0)[0]

        if version is not _V0:
            raise FrugalVersionException(
                "Wrong Frugal version.  Found version {0}.  Wanted version {1}"
                .format(version, _V0))

        size = unpack_from(_UINT, buff, 1)[0]

        offset = 5  # since size is 4 bytes

        while offset < size:
            key_size = unpack_from(_UINT, buff, offset)[0]
            offset += 4

            key = unpack_from(_string(key_size), buff, offset)[0]
            offset += len(key)

            val_size = unpack_from(_UINT, buff, offset)[0]
            offset += 4

            val = unpack_from(_string(val_size), buff, offset)[0]
            offset += len(val)

            parsed_headers[key] = val

        return parsed_headers


def _string(self, length):
    return '>{}s'.format(length)
