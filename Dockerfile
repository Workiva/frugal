FROM amazonlinux:2 AS build-go
SHELL ["/bin/bash", "-c"]

# Pin deps versions
ARG ACTIVEMQ_VER=5.15.6
ARG DART_VER=2.7.1
ARG GO_VER=1.14.2
ARG MAVEN_VER=3.3.9

WORKDIR /build/
ADD . /build/

# Copy the NATS binary
COPY --from=nats:2.1.2-linux /nats-server /usr/local/bin/nats-server

ENV GOPATH /go
ENV JAVA_HOME /usr/lib/jvm/java-1.8.0-openjdk/
ENV JAVA java
ENV JAVAC javac
ENV PATH /opt/maven/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/bin:/usr/lib:/usr/local/go/bin:/opt/google/dart/bin:/go/bin:/opt/activemq/bin

# Install our plethora of yum packages
ARG BUILD_ID
RUN yum update -y && \
    yum groupinstall -y "Development Tools" && \
    yum install -y \
        ant \
        binutils \
        curl \
        git \
        gtk2-devel \
        java-1.8.0-openjdk \
        libffi-devel \
        mercurial \
        openssl-devel \
        procps \
        python-devel \
        python-pip \
        python-setuptools \
        python3-devel \
        python3-pip \
        tar \
        unzip \
        wget \
        xorg-x11-server-Xvfb.x86_64 \
        xz \
        zip && \
    yum upgrade -y && \
    yum autoremove -y && \
    yum clean all && \
    rm -rf /var/cache/yum

# Install ActiveMQ 5.15.6
RUN wget -q http://archive.apache.org/dist/activemq/${ACTIVEMQ_VER}/apache-activemq-${ACTIVEMQ_VER}-bin.tar.gz && \
    tar -xzf apache-activemq-${ACTIVEMQ_VER}-bin.tar.gz -C /opt && \
    mv /opt/apache-activemq* /opt/activemq && \
    chmod -R 755 /opt/activemq && \
    rm apache-activemq-${ACTIVEMQ_VER}-bin.tar.gz

# Install dart
RUN curl -s https://storage.googleapis.com/dart-archive/channels/stable/release/${DART_VER}/sdk/dartsdk-linux-x64-release.zip > \
    /dartsdk.zip && \
    unzip -qq /dartsdk.zip -d /opt/google && \
    mv /opt/google/dart-sdk* /opt/google/dart && \
    chmod -R 775 /opt/google/dart && \
    rm /dartsdk.zip

# install chrome
RUN wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm && \
    yum install -y ./google-chrome-stable_current_*.rpm && yum clean all && rm -rf /var/cache/yum && \
    rm ./google-chrome-stable_current_*.rpm && \
    mv /usr/bin/google-chrome-stable /usr/bin/google-chrome && \
    sed -i --follow-symlinks -e 's/\"\$HERE\/chrome\"/\"\$HERE\/chrome\" --no-sandbox/g' /usr/bin/google-chrome && \
    google-chrome --version

# Install maven
RUN curl -s http://mirrors.ibiblio.org/apache/maven/maven-3/${MAVEN_VER}/binaries/apache-maven-${MAVEN_VER}-bin.zip > \
    /maven.zip && \
    unzip /maven.zip -d /opt/ && \
    mv /opt/apache-maven* /opt/maven && \
    chmod -R 755 /opt/maven && \
    rm /maven.zip

# Install go
RUN curl -so /tmp/go${GO_VER}.linux-amd64.tar.gz \
    https://storage.googleapis.com/golang/go${GO_VER}.linux-amd64.tar.gz && \
    tar -C /usr/local/ -xzf /tmp/go${GO_VER}.linux-amd64.tar.gz && \
    rm -Rf /tmp/go${GO_VER}.linux-amd64.tar.gz && \
    rm -r /usr/local/go/doc /usr/local/go/api

# Who doesn't love virtualenv
RUN pip3 install virtualenv
# Install jsonlint
RUN pip3 install demjson

# Adudit pip artifacts (possibly not even needed anymore)
RUN mkdir /audit/
ARG BUILD_ARTIFACTS_AUDIT=/audit/*
RUN pip3 freeze > /audit/pip.lock

CMD ["bash"]
# Update packages
ARG BUILD_ID
RUN yum update -y && \

ARG GIT_BRANCH
ARG GIT_MERGE_BRANCH
ARG GIT_SSH_KEY
ARG KNOWN_HOSTS_CONTENT

ARG BUILD_ID
RUN yum update -y && \
    yum upgrade -y && \
    yum clean all && \
    rm -rf /var/cache/yum

# the `sed` cmd is to enable the `--no-sandbox` flag when running chrome, which is necessary to run dart unit tests in the browser.
RUN sed -i --follow-symlinks -e 's/\"\$HERE\/chrome\"/\"\$HERE\/chrome\" --no-sandbox/g' /usr/bin/google-chrome

ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/messaging-sdk/python2_pip_deps.txt:/go/src/github.com/Workiva/messaging-sdk/python3_pip_deps.txt:/go/src/github.com/Workiva/messaging-sdk/lib/go/go.mod:/go/src/github.com/Workiva/messaging-sdk/lib/dart/sdk/pubspec.lock:/go/src/github.com/Workiva/messaging-sdk/lib/java/sdk/pom.xml
ARG BUILD_ARTIFACTS_DOCUMENTATION=/go/src/github.com/Workiva/messaging-sdk/lib/dart/sdk/doc/api/api.tar.gz
ARG BUILD_ARTIFACTS_PYPI=/go/src/github.com/Workiva/messaging-sdk/messaging-sdk-*.tar.gz
ARG BUILD_ARTIFACTS_ARTIFACTORY=/go/src/github.com/Workiva/messaging-sdk/messaging-sdk-*.jar
ARG BUILD_ARTIFACTS_PUB=/go/src/github.com/Workiva/messaging-sdk/messaging_sdk.pub.tgz

ARG PIP_INDEX_URL
ARG PIP_EXTRA_INDEX_URL=https://pypi.python.org/simple/

ARG ARTIFACTORY_PRO_USER
ARG ARTIFACTORY_PRO_PASS

ARG GIT_BRANCH
ARG GIT_MERGE_BRANCH
ARG GIT_SSH_KEY
ARG KNOWN_HOSTS_CONTENT

ARG GOPATH=/go/
ENV PATH $GOPATH/bin:$PATH
ENV MESSAGING_SDK_HOME=/go/src/github.com/Workiva/messaging-sdk
ENV GOPRIVATE=github.com/Workiva

RUN git config --global url.git@github.com:.insteadOf https://github.com
RUN mkdir /root/.ssh && \
    echo "$KNOWN_HOSTS_CONTENT" > "/root/.ssh/known_hosts" && \
    chmod 700 /root/.ssh/ && \
    umask 0077 && echo "$GIT_SSH_KEY" >/root/.ssh/id_rsa && \
    eval "$(ssh-agent -s)" && ssh-add /root/.ssh/id_rsa
ADD ./settings.xml /root/.m2/settings.xml

WORKDIR /go/src/github.com/Workiva/messaging-sdk/
ADD . /go/src/github.com/Workiva/messaging-sdk/

RUN echo "Starting the script section" && \
		./scripts/smithy/smithy.sh && \
		cat $MESSAGING_SDK_HOME/test_results/smithy_dart.sh_out.txt && \
		cat $MESSAGING_SDK_HOME/test_results/smithy_go.sh_out.txt && \
		cat $MESSAGING_SDK_HOME/test_results/smithy_python_two.sh_out.txt && \
		cat $MESSAGING_SDK_HOME/test_results/smithy_python_three.sh_out.txt && \
		cat $MESSAGING_SDK_HOME/test_results/smithy_java.sh_out.txt && \
		echo "script section completed"

ARG GIT_BRANCH
ARG GIT_MERGE_BRANCH
ARG GIT_SSH_KEY
ARG KNOWN_HOSTS_CONTENT
WORKDIR /go/src/github.com/Workiva/frugal/
ADD . /go/src/github.com/Workiva/frugal/

RUN mkdir /root/.ssh && \
    echo "$KNOWN_HOSTS_CONTENT" > "/root/.ssh/known_hosts" && \
    chmod 700 /root/.ssh/ && \
    umask 0077 && echo "$GIT_SSH_KEY" >/root/.ssh/id_rsa && \
    eval "$(ssh-agent -s)" && ssh-add /root/.ssh/id_rsa

ARG BUILD_ID
RUN yum update -y && \
    yum upgrade -y && \
    yum clean all && \
    rm -rf /var/cache/yum

ARG GOPATH=/go/
ENV PATH $GOPATH/bin:$PATH
RUN git config --global url.git@github.com:.insteadOf https://github.com
ENV FRUGAL_HOME=/go/src/github.com/Workiva/frugal
RUN echo "Starting the script section" && \
		./scripts/smithy.sh && \
		cat $FRUGAL_HOME/test_results/smithy_dart.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_go.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_generator.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_python.sh_out.txt && \
		cat $FRUGAL_HOME/test_results/smithy_java.sh_out.txt && \
		echo "script section completed"

ARG BUILD_ARTIFACTS_RELEASE=/go/src/github.com/Workiva/frugal/frugal
ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/frugal/python2_pip_deps.txt:/go/src/github.com/Workiva/frugal/python3_pip_deps.txt:/go/src/github.com/Workiva/frugal/lib/go/go.mod:/go/src/github.com/Workiva/frugal/lib/dart/pubspec.lock:/go/src/github.com/Workiva/frugal/lib/java/pom.xml
ARG BUILD_ARTIFACTS_GO_LIBRARY=/go/src/github.com/Workiva/frugal/goLib.tar.gz
ARG BUILD_ARTIFACTS_PYPI=/go/src/github.com/Workiva/frugal/frugal-*.tar.gz
ARG BUILD_ARTIFACTS_ARTIFACTORY=/go/src/github.com/Workiva/frugal/frugal-*.jar
ARG BUILD_ARTIFACTS_PUB=/go/src/github.com/Workiva/frugal/frugal.pub.tgz
ARG BUILD_ARTIFACTS_TEST_RESULTS=/go/src/github.com/Workiva/frugal/test_results/*

# make a simple etc/passwd file so the scratch image can run as nobody (not root)
RUN echo "nobody:x:65534:65534:Nobody:/:" > /passwd.minimal

FROM scratch
COPY --from=build /go/src/github.com/Workiva/frugal/frugal /bin/frugal
COPY --from=build /passwd.minimal /etc/passwd
USER nobody
ENTRYPOINT ["frugal"]
