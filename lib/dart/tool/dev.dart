library tool.dev;

import 'dart:async';
import 'package:dart_dev/dart_dev.dart'
    show Environment, dev, TestRunnerConfig, config;

Future main(List<String> args) async {
  // https://github.com/Workiva/dart_dev

  var directories = ['lib/', 'test/', 'tool/'];
  config.analyze.entryPoints = directories;
  config.format.directories = directories;

  var genTest = "test/generated_runner_test.dart";

  config.test
    ..platforms = ['vm']
    ..unitTests = [genTest];

  config.genTestRunner.configs = [
    new TestRunnerConfig(
        directory: 'test',
        env: Environment.vm,
        filename: 'generated_runner_test')
  ];

  config.format..exclude = [genTest];

  await dev(args);
}
