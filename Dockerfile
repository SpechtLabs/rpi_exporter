# Build the rpi_exporter binary. The builder runs on the build machine's own
# platform and cross-compiles for the target, rather than compiling under
# QEMU emulation. Keep the golang tag in lockstep with go in .mise.toml.
FROM --platform=$BUILDPLATFORM golang:1.27.2 AS builder

ARG TARGETOS
ARG TARGETARCH
# The release version, which `rpi_exporter --version` prints.
ARG VERSION=dev

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY rpi_exporter.go rpi_exporter.go
COPY collector/ collector/

# Build
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
    -ldflags "-s -w -X=github.com/prometheus/common/version.Version=${VERSION#v}" \
    -o rpi_exporter .

FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Lukas Malkmus <mail@lukasmalkmus.com>"

COPY --from=builder /workspace/rpi_exporter /bin/rpi_exporter

ENTRYPOINT ["/bin/rpi_exporter"]
EXPOSE     9243
