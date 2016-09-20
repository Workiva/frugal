# frugal-java

Java library for [Frugal](https://github.com/Workiva/frugal).


## Running Tests

Maven is configured to compile a Frugal IDL for use in integration tests.
In your IDE, you will need to add `target/generated-test-sources/frugal`
as a source folder to your project.

Integration tests are run during the mvn `integration-test` phase (after
packaging).

To run all unit tests and integration tests.

```bash
mvn verify
or
mvn install
```

To skip integration tests.

```bash
mvn -DskipITs
```

To skip all tests.

```bash
mvn -DskipTests
```
