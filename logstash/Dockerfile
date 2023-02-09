ARG LOGSTASH_VERSION
FROM logstash:${LOGSTASH_VERSION}

# Remove jvm heap parameters from jvm.options file to allow setting them using env vars
ARG JVM_OPTS_FILE=/usr/share/logstash/config/jvm.options
RUN sed -r -i 's/^-Xm[xs].+$//' ${JVM_OPTS_FILE}

RUN /opt/logstash/bin/logstash-plugin install logstash-integration-jdbc

RUN mkdir /opt/logstash/vendor/jdbc

ARG JDBC_POSTGRES_DRIVER_VERSION
RUN curl -o /opt/logstash/vendor/jdbc/postgresql.jar https://jdbc.postgresql.org/download/postgresql-${JDBC_POSTGRES_DRIVER_VERSION}.jar

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]