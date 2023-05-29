
clean-local:
	mvn clean

deps-local:
	@:  # pass

check-local: clean
	mvn checkstyle:checkstyle

test-local: clean
	mvn verify -q -Dcheckstyle.skip
