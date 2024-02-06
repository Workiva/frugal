FROM drydock-prod.workiva.net/workiva/messaging-docker-images:pr-86 as build

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
RUN java -version
ARG BUILD_ID
ARG GOPATH=/go/
ENV PATH $GOPATH/bin:$PATH

#### Run maven deps
ARG ARTIFACTORY_PRO_USER
ARG ARTIFACTORY_PRO_PASS
ENV MAVEN_ROOT /go/src/github.com/Workiva/frugal/lib/java

RUN git config --global url.git@github.com:.insteadOf https://github.com
ENV FRUGAL_HOME=/go/src/github.com/Workiva/frugal
RUN echo "Starting the script section"
RUN ./scripts/smithy.sh
RUN cat $FRUGAL_HOME/test_results/smithy_dart.sh_out.txt
RUN cat $FRUGAL_HOME/test_results/smithy_go.sh_out.txt
RUN cat $FRUGAL_HOME/test_results/smithy_generator.sh_out.txt
RUN cat $FRUGAL_HOME/test_results/smithy_python.sh_out.txt
RUN cat $FRUGAL_HOME/test_results/smithy_java.sh_out.txt
RUN echo "script section completed"

ARG BUILD_ARTIFACTS_RELEASE=/go/src/github.com/Workiva/frugal/frugal
ARG BUILD_ARTIFACTS_AUDIT=/go/src/github.com/Workiva/frugal/python2_pip_deps.txt:/go/src/github.com/Workiva/frugal/python3_pip_deps.txt:/go/src/github.com/Workiva/frugal/lib/go/go.mod:/go/src/github.com/Workiva/frugal/lib/dart/pubspec.lock:/go/src/github.com/Workiva/frugal/lib/java/pom.xml
ARG BUILD_ARTIFACTS_GO_LIBRARY=/go/src/github.com/Workiva/frugal/goLib.tar.gz
ARG BUILD_ARTIFACTS_PYPI=/go/src/github.com/Workiva/frugal/frugal-*.tar.gz
ARG BUILD_ARTIFACTS_JAVA=/go/src/github.com/Workiva/frugal/frugal-*.jar
ARG BUILD_ARTIFACTS_PUB=/go/src/github.com/Workiva/frugal/frugal.pub.tgz
ARG BUILD_ARTIFACTS_TEST_RESULTS=/go/src/github.com/Workiva/frugal/test_results/*

# make a simple etc/passwd file so the scratch image can run as nobody (not root)
RUN echo "nobody:x:65534:65534:Nobody:/:" > /passwd.minimal

FROM scratch
COPY --from=build /go/src/github.com/Workiva/frugal/frugal /bin/frugal
COPY --from=build /passwd.minimal /etc/passwd
USER nobody
ENTRYPOINT ["frugal"]
