library tool.dev;

import 'package:dart_dev/dart_dev.dart'
    show Environment, dev, TestRunnerConfig, config;

main(List<String> args) async {
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
