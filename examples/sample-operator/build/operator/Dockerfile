FROM docker.io/golang:1.22 AS builder
WORKDIR /src/sample-operator/
COPY . .
ENV GOFLAGS=-mod=vendor
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o /sample-operator cmd/manager/main.go

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/sample-operator \
    USER_UID=1001 \
    USER_NAME=sample-operator

# install operator binary
COPY --from=builder /sample-operator ${OPERATOR}
COPY build/operator/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
