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

FROM --platform=linux/${TARGETARCH} scratch

WORKDIR /opt/vcloud/bin

COPY --from=builder /build/vcloud/cluster-api-provider-cloud-director .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# nobody user ID
USER 65534
ENTRYPOINT ["/opt/vcloud/bin/cluster-api-provider-cloud-director"]
