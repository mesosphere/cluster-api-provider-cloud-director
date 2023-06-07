# syntax=docker/dockerfile:1

# Build the manager binary
FROM --platform=linux/$BUILDARCH golang:1.19 as builder

RUN apt-get update && \
    apt-get -y install \
        bash \
        git  \
        make

ADD . /go/src/github.com/vmware/cluster-api-provider-cloud-director
WORKDIR /go/src/github.com/vmware/cluster-api-provider-cloud-director

ENV GOPATH /go
ENV GOARCH $TARGETARCH
ARG VERSION
RUN make build-within-docker VERSION=$VERSION && \
    chmod +x /build/vcloud/cluster-api-provider-cloud-director

########################################################

# Use distroless/static:nonroot image for a base.
FROM --platform=linux/amd64 gcr.io/distroless/static@sha256:1b4dbd7d38a0fd4bbaf5216a21a615d07b56747a96d3c650689cbb7fdc412b49 as linux-amd64
FROM --platform=linux/arm64 gcr.io/distroless/static@sha256:05810557ec4b4bf01f4df548c06cc915bb29d81cb339495fe1ad2e668434bf8e as linux-arm64

# Build the actual final image
FROM --platform=linux/${TARGETARCH} linux-${TARGETARCH}

WORKDIR /opt/vcloud/bin

COPY --from=builder /build/vcloud/cluster-api-provider-cloud-director .

# nobody user ID
USER 65534
ENTRYPOINT ["/opt/vcloud/bin/cluster-api-provider-cloud-director"]
