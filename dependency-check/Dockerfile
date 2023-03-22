FROM owasp/dependency-check:8.1.2

COPY bkrepo-dependency-check /bkrepo-dependency-check

USER root
RUN chown dependencycheck:0 /bkrepo-dependency-check && \
    chmod u+x,g=u /bkrepo-dependency-check && \
    mkdir /bkrepo && \
    chown -R dependencycheck:0 /bkrepo && \
    chmod -R g=u /bkrepo

USER dependencycheck
CMD []
ENTRYPOINT []
