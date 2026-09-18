# Built from a prebuilt binary (see `make build-linux`), same reasoning as Dockerfile:
# this environment's Docker network has flaky/slow HTTPS to proxy.golang.org.
FROM alpine:3.20
COPY bin/openmeter-mock-linux-amd64 /usr/local/bin/openmeter-mock
ENV OPENMETER_MOCK_HTTP_PORT=9010
ENV OPENMETER_MOCK_MONGO_URI=mongodb://mongo:27017
ENV OPENMETER_MOCK_MONGO_DB=openmeter_mock
EXPOSE 9010
ENTRYPOINT ["/usr/local/bin/openmeter-mock"]
