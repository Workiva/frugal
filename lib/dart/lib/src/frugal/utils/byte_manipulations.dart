import 'dart:typed_data';

int readInt(Uint8List buff, int i) {
  return ((buff[i] & 0xff) << 24) |
  ((buff[i + 1] & 0xff) << 16) |
  ((buff[i + 2] & 0xff) << 8) |
  (buff[i + 3] & 0xff);
}

void writeInt(int i, Uint8List buff, int i1) {
  buff[i1] = (0xff & (i >> 24));
  buff[i1 + 1] = (0xff & (i >> 16));
  buff[i1 + 2] = (0xff & (i >> 8));
  buff[i1 + 3] = (0xff & (i));
}

void writeStringBytes(List<int> strBytes, Uint8List buff, int i) {
  buff.setRange(i, i + strBytes.length, strBytes);
}
