from frugal.util.headers import _Headers


def mock_frame(context):
    return _Headers._write_to_bytearray(context.get_request_headers())

