# Built from a prebuilt binary (see scripts/build-linux.sh / `make docker-build`) rather than
# compiling inside a multi-stage build — this environment's Docker network has flaky/slow HTTPS
# to proxy.golang.org, and the host already has a working Go toolchain.
FROM alpine:3.20
COPY bin/llm-mock-api-linux-amd64 /usr/local/bin/llm-mock-api
COPY admin/scenarios.example.yaml /etc/llm-mock-api/scenarios.yaml
ENV MOCK_HTTP_PORT=9000
ENV MOCK_ADMIN_PORT=9001
ENV MOCK_SCENARIO_FILE=/etc/llm-mock-api/scenarios.yaml
EXPOSE 9000 9001
ENTRYPOINT ["/usr/local/bin/llm-mock-api"]
