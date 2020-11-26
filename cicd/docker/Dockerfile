# Copyright 2020 program was created by VMware, Inc.
# SPDX-License-Identifier: Apache-2.0

# Intermediate build container
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Copy
ADD . /app/

# Build
WORKDIR /app
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o resource cmd/main.go

# Copy
RUN mkdir -p /target/opt/resource/ && \
    cp resource /target/opt/resource/

# Create links for in, out and check commands
# pointing to resource binary
RUN ln -s resource /target/opt/resource/in && \
    ln -s resource /target/opt/resource/out && \
    ln -s resource /target/opt/resource/check

# Final container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /target/opt /opt