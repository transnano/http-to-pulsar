# Start by building the application.
FROM golang:1.19.2-buster as build

WORKDIR /go/src/github.com/transnano/http-to-pulsar/
# For building Go Module required
ENV GOPROXY=direct
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  go mod download
# Copy the go source
COPY . .
# Build
RUN  go build -a -o main -ldflags "-s -w"

# Now copy it into our base image.
# hadolint ignore=DL3006
FROM gcr.io/distroless/base-debian10
#FROM gcr.io/distroless/base
LABEL maintainer="Transnano <transnano.jp@gmail.com>"
COPY --from=build /go/src/github.com/transnano/http-to-pulsar/main /
ENTRYPOINT ["/main"]
