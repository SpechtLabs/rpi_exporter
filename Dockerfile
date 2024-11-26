# Build the urlshortener binary
FROM golang:1.23 AS builder

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
RUN go build -o rpi_exporter ./rpi_exporter.go

FROM  quay.io/prometheus/busybox:latest
LABEL maintainer="Lukas Malkmus <mail@lukasmalkmus.com>"

COPY --from=builder /workspace/rpi_exporter /bin/rpi_exporter

ENTRYPOINT ["/bin/rpi_exporter"]
EXPOSE     9243
