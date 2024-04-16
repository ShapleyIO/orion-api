FROM golang:1.22.1 as base

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts \
    git config --global url."git@github.com".insteadOf "https://github.com"

ENV PROJECT shapley-shapley.io-api
WORKDIR $GOPATH/src/github.com/ShapleyIO/$PROJECT
RUN git config --global --add safe.directory $GOPATH/src/github.com/ShapleyIO/$PROJECT

FROM base as builder

RUN apt-get update && apt-get install --yes --quiet \
    netcat-openbsd unzip \
    && rm -rf /var/lib/apt/lists/*

ENV OAPI_CODEGEN_VERSION 1.16.2
ENV MOCKGEN_VERSION v1.6.0

VOLUME $GOPATH/src/github.com/ShapleyIO/$PROJECT

# Install Tools
COPY go.mod ./go.mod
COPY go.sum ./go.sum
COPY vendor vendor

RUN GOFLAGS='' go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v${OAPI_CODEGEN_VERSION}
RUN GOFLAGS='' go install github.com/golang/mock/mockgen@${MOCKGEN_VERSION}

FROM base as golang-builder
ARG BUILD_DATE
ARG REVISION
ARG OAPI_CODEGEN_VERSION
COPY . .

FROM golang-builder as api-builder
RUN --mount=type=cache,target=/root/.cache/go-build bin/build

FROM golang:1.22.1 as api
COPY --from=api-builder /go/src/github.com/ShapleyIO/shapley-shapley.io-api/dist/shapley.io-api /usr/bin/shapley.io-api
ENV INTERFACE="[::]"
EXPOSE 8080
ENTRYPOINT [ "shapley.io-api" ]
