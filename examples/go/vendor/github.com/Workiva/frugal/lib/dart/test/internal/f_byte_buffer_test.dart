import "dart:typed_data";
import "package:test/test.dart";

import "../../lib/src/frugal.dart";

var _list = [
  0,
  0,
  0,
  0,
  29,
  0,
  0,
  0,
  3,
  102,
  111,
  111,
  0,
  0,
  0,
  3,
  98,
  97,
  114,
  0,
  0,
  0,
  4,
  98,
  108,
  97,
  104,
  0,
  0,
  0,
  3,
  98,
  97,
  122
];

void main() {
  test('test that write properly writes the bytes from the given buffer', () {
    var buffList = new Uint8List.fromList(_list);
    var buff = new FByteBuffer(10);
    expect(10, buff.writeRemaining);
    var n = buff.write(buffList, 0, buffList.length);
    expect(n, 10);
    var expected = new Uint8List.fromList(_list.sublist(0, 10));
    expect(buff.asUint8List(), expected);
    expect(0, buff.writeRemaining);
  });

  test('test that read properly reads the bytes into the given buffer', () {
    var buffList = new Uint8List.fromList(_list);
    var buff = new FByteBuffer.fromUint8List(buffList);
    var readBuff = new Uint8List(10);
    expect(_list.length, buff.readRemaining);
    var n = buff.read(readBuff, 0, 15);
    expect(10, n);
    var expected = new Uint8List.fromList(_list.sublist(0, 10));
    expect(readBuff, expected);
    expect(_list.length - 10, buff.readRemaining);
  });
}
