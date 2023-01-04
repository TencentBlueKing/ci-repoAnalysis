FROM aquasec/trivy:0.34.0

COPY trivy /trivy
COPY metadata.json /root/.cache/trivy/db
RUN mkdir -p /root/.cache/trivy/db && chmod +x /trivy
ENTRYPOINT []
