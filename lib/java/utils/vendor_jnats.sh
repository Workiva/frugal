#!/bin/bash

set -e

mkdir -p lib/io/nats/jnats/0.4.0-wk
curl https://oss.sonatype.org/content/groups/public/io/nats/jnats/0.4.0-SNAPSHOT/jnats-0.4.0-20160304.053128-17.pom | sed 's@<version>0.4.0-SNAPSHOT</version>@<version>0.4.0-wk</version>@g' > lib/io/nats/jnats/0.4.0-wk/jnats-0.4.0-wk.pom
curl https://oss.sonatype.org/content/groups/public/io/nats/jnats/0.4.0-SNAPSHOT/jnats-0.4.0-20160304.053128-17.jar > lib/io/nats/jnats/0.4.0-wk/jnats-0.4.0-wk.jar
