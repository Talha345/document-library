#
#  Author: Hari Sekhon
#  Date: 2016-04-24 21:18:57 +0100 (Sun, 24 Apr 2016)
#
#  vim:ts=4:sts=4:sw=4:et
#
#  https://github.com/harisekhon/Dockerfiles/hbase
#
#  If you're using my code you're welcome to connect with me on LinkedIn and optionally send me feedback
#
#  https://www.linkedin.com/in/harisekhon
#

FROM alpine:latest
MAINTAINER Hari Sekhon (https://www.linkedin.com/in/harisekhon)
# HBASE version updated to stable version as previousy used link was not available anymore.
ARG HBASE_VERSION=2.1.0

ENV PATH $PATH:/hbase/bin

ENV JAVA_HOME=/usr

LABEL Description="HBase Dev", \
      "HBase Version"="$HBASE_VERSION"

WORKDIR /

RUN \
    apk add --no-cache bash openjdk8-jre-base wget tar openssh && \
    url="https://archive.apache.org/dist/hbase/$HBASE_VERSION/hbase-$HBASE_VERSION-bin.tar.gz"; \
    wget -t 100 --retry-connrefused -O "hbase-$HBASE_VERSION-bin.tar.gz" "$url" && \
    mkdir hbase-$HBASE_VERSION && \
    tar zxf hbase-$HBASE_VERSION-bin.tar.gz -C hbase-$HBASE_VERSION --strip 1 && \
    ln -sv hbase-$HBASE_VERSION hbase && \
    rm -fv hbase-$HBASE_VERSION-bin.tar.gz && \
    { rm -rf hbase/{docs,src}; : ; } && \
    apk del wget tar

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

# Needed for HBase 2.0+ hbase-shell
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && \
    apk add jruby && \
    apk add jruby-irb # jruby-maven jruby-minitest jruby-rdoc jruby-rake jruby-testunit

COPY entrypoint.sh /
RUN chmod +x entrypoint.sh
COPY conf/hbase-site.xml /hbase/conf/
COPY profile.d/java.sh /etc/profile.d/
COPY hbase_init.txt /

# Stargate  8080  / 8085
# Thrift    9090  / 9095
# HMaster   16000 / 16010
# RS        16201 / 16301
EXPOSE 2181 8080 8085 9090 9095 16000 16010 16201 16301

CMD "/entrypoint.sh"
