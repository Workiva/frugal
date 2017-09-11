# Protocol

This describes the binary protocol used to encode FContext by an FProtocol.

FProtocol serializes FContext headers using a custom protocol before the normal
serialization of the Thrift message, as produced by TProtocol. FProtocol is a
framed protocol, meaning the length of the serialized message, or frame, is
prepended to the frame itself. As such, a serialized Frugal message looks like
the following on the wire at a high level:

```
+------------+------------------+-------------------+
| frame size | FContext headers | TProtocol message |
+------------+------------------+-------------------+
```

The serialization of the TProtocol message is handled entirely by the Thrift
TProtocol. For example, this could itself be framed if a TFramedTransport is
used. However, the frame size and FContext headers are serialized by FProtocol.
The header protocol reserves a single byte for versioning purposes. Currently,
only v0 is supported.

The complete binary wire layout is documented below. Network byte order is
assumed.

```
   0     1     2     3     4     5     6     7     8     9     10    11    12    13    14  ...
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+...+-----+-----+-----+-----+-----+-----+-----+...+-----+...+-----+-----+...+-----+
|     frame size n      | ver |    headers size m     |  header name size k   |  0  |  1  |...| k-1 |  header value size v  |  0  |  1  |...| v-1 |...|  0  |  1  |...| t-1 |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+...+-----+-----+-----+-----+-----+-----|-----+...+-----+...+-----+-----+...+-----+
|<-------4 bytes------->|<----------5 bytes---------->|<-------4 bytes------->|<------k bytes------>|<-------4 bytes------->|<------v bytes------>|   |<------t bytes------>|
                                                      |<-------------------------------------------m bytes------------------------------------------->|
|<------------------------------------------------------------------------------n bytes------------------------------------------------------------------------------------>|
```

| Name                | Size    | Definition                                                   |
|---------------------|---------|--------------------------------------------------------------|
| frame size n        | 4 bytes | unsigned integer representing length of entire frame         |
| ver                 | 1 byte  | unsigned integer representing header protocol version        |
| headers size m      | 4 bytes | unsigned integer representing length of header data          |
| header name size k  | 4 bytes | unsigned integer representing the length of the header name  |
| header name         | k bytes | the header name                                              |
| header value size v | 4 bytes | unsigned integer representing the length of the header value |
| header value        | v bytes | the header value                                             |
| Thrift message      | t bytes | the TProtocol-serialized message                             |
Header key-value pairs are repeated
