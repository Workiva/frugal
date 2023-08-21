import 'package:dart_dev/dart_dev.dart';
import 'package:dart_dev_workiva/dart_dev_workiva.dart';
import 'package:over_react_format/dart_dev_tool.dart';

final Map<String, DevTool> config = {
  ...workivaConfig,
  'test': WorkivaTestTool(),
  'format': OverReactFormatTool()
    ..organizeDirectives = true
    ..lineLength = 120,
};
