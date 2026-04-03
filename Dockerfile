# Copyright 2026 Chalindu Kodikara
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /licenseops ./cmd/licenseops/

FROM alpine:3.21

COPY --from=builder /licenseops /usr/local/bin/licenseops

ENTRYPOINT ["licenseops"]
