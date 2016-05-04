from struct import pack_into, unpack_from, unpack

from frugal.exceptions import FrugalVersionException, FProtocolException


_V0 = 0
# Code for big endian unsigned char
_UCHAR = '!B'
# Code for big endian unsigned int
_UINT = '!I'
_UINT_LENGTH = 4


class _Headers(object):

    @staticmethod
    def _write_to_bytearray(headers):
        """Writes a given dictionary to a bytearray object and returns it
        TODO : Point to Protocol doc in the repo

        Args:
            headers: dict of frugal headers to write
        Returns:
            bytearray containing binary headers
        """
        size = 0

        for key, value in headers.iteritems():
            size = size + 8 + len(key) + len(value)

        buff = bytearray(size + 5)

        pack_into(_UCHAR, buff, 0, _V0)
        pack_into(_UINT, buff, 1, size)

        offset = 5

        for key, value in headers.iteritems():
            key_length = len(key)
            pack_into(_UINT, buff, offset, key_length)
            offset += 4

            pack_into('>{0}s'.format(str(key_length)), buff, offset, key)
            offset += len(key)

            pack_into(_UINT, buff, offset, len(value))
            offset += 4

            pack_into('>{0}s'.format(str(len(value))), buff, offset, value)
            offset += len(value)

        return buff

    @staticmethod
    def _read(buff1):
        buff = buff1.read(1)
        parsed_headers = {}
        version = unpack_from(_UCHAR, buff[:1])[0]
        # TODO: Check the version!

        buff = buff1.read(4)
        size = unpack_from(_UINT, buff[0:4])[0]

        offset = 0  # since size is 4 bytes

        buff = buff1.read(size)
        while offset < size:
            key_size = unpack_from(_UINT, buff[offset:offset + 4])[0]
            offset += 4

            # TODO: Check bounds.

            key = unpack_from('>{0}s'.format(key_size), buff[offset:offset+key_size])[0]
            offset += key_size

            # TODO: Check bounds.

            val_size = unpack_from(_UINT, buff, offset)[0]
            offset += 4

            # TODO: Check bounds.

            val = unpack_from('>{0}s'.format(val_size), buff, offset)[0]
            offset += val_size
            parsed_headers[key] = val

        return parsed_headers

    @staticmethod
    def read(transport):
        """ Read frugal frame from TTranpsort

        Args:
            transport: TTransport containing frugal frame
        """
        #buff = transport.readAll(4)
        #frugal_frame_size = unpack(!I, buff)
        pass

    @staticmethod
    def decode_from_frame(frame):
        if len(frame) < 5:
            raise FProtocolException(
                FProtocolException.INVALID_DATA,
                "Invalid frame size: {}".format(len(frame))
            )

        # TODO: What do we do with this?
        unpack_from('!I', frame[:4])[0]

        version = unpack_from(_UCHAR, frame[4:5])[0]

        if version is not _V0:
            raise FProtocolException(
                FProtocolException.BAD_VERSION,
                "Wrong Frugal version. Found {0}, wanted {1}."
                .format(version, _V0))

        headers_size = unpack_from('!I', frame[5:9])[0]

        return _Headers._read_pairs(frame, 9, headers_size + 9)

    @staticmethod
    def _read_pairs(buff, start, end):
        parsed_headers = {}
        i = start
        while i < end:
            name_size = unpack_from(_UINT, buff[i:i + 4])[0]
            i += 4

            if i > end or i + name_size > end:
                raise FProtocolException(FProtocolException.INVALID_DATA,
                                         "invalid protocol header name size: {}"
                                         .format(name_size))

            name = unpack_from('>{0}s'.format(name_size),
                               buff[i:i + name_size])[0]
            i += name_size

            val_size = unpack_from(_UINT, buff[i: i + 4])[0]
            i += 4

            if i > end or i + val_size > end:
                raise FProtocolException(
                    FProtocolException.INVALID_DATA,
                    "invalid protocol header value size: {}".format(val_size)
                )

            val = unpack_from('>{0}s'.format(val_size), buff[i:i + val_size])[0]
            i += val_size
            parsed_headers[name] = val

        return parsed_headers


