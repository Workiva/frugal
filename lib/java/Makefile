clean-local:
	mvn clean

deps-local:
	@:  # pass

check-local: clean-local
	mvn checkstyle:checkstyle

test-local: clean-local
	mvn verify -q -Dcheckstyle.skip
