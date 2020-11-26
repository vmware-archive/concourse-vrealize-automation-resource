# Copyright 2020 program was created by VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Copy
ADD . /app/

# Build
WORKDIR /app
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o resource cmd/main.go
RUN mkdir -p /target/opt/resource/
RUN cp resource /target/opt/resource/

# Create links for in, out and check commands
# pointing to resource binary
RUN ln -s resource /target/opt/resource/in
RUN ln -s resource /target/opt/resource/out
RUN ln -s resource /target/opt/resource/check


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /target/opt /opt