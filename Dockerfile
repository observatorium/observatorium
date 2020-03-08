# Build the manager binary
FROM golang:1.13.3-alpine3.10 as builder

RUN apk add --update --no-cache git bash
WORKDIR /workspace
# Copy the jsonnet source
COPY environments/operator/ environments/operator/
COPY components/ components/
COPY jsonnetfile.json jsonnetfile.json
# Build
RUN GO111MODULE="on" go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
RUN jb install
RUN GO111MODULE="on" go get github.com/brancz/locutus

FROM alpine:3.10 as runner
WORKDIR /
COPY --from=builder /go/bin/locutus /
COPY --from=builder /workspace/environments/operator /environments/operator
COPY --from=builder /workspace/components/ /components/
COPY --from=builder /workspace/vendor/ /vendor/
RUN chgrp -R 0 /vendor && chmod -R g=u /vendor
RUN chgrp -R 0 /environments && chmod -R g=u /environments
RUN chgrp -R 0 /components && chmod -R g=u /components
ENTRYPOINT ["/locutus", "--renderer=jsonnet", "--renderer.jsonnet.entrypoint=environments/operator/main.jsonnet", "--trigger=resource", "--trigger.resource.config=environments/operator/config.yaml"]
