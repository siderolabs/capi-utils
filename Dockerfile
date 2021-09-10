# syntax = docker/dockerfile-upstream:1.2.0-labs

# THIS FILE WAS AUTOMATICALLY GENERATED, PLEASE DO NOT EDIT.
#
# Generated on 2021-09-10T15:47:01Z by kres latest.

ARG TOOLCHAIN

# cleaned up specs and compiled versions
FROM scratch AS generate

FROM ghcr.io/talos-systems/ca-certificates:v0.3.0-12-g90722c3 AS image-ca-certificates

FROM ghcr.io/talos-systems/fhs:v0.3.0-12-g90722c3 AS image-fhs

# runs markdownlint
FROM node:14.8.0-alpine AS lint-markdown
RUN npm i -g markdownlint-cli@0.23.2
RUN npm i sentences-per-line@0.2.1
WORKDIR /src
COPY .markdownlint.json .
COPY ./README.md ./README.md
RUN markdownlint --ignore "CHANGELOG.md" --ignore "**/node_modules/**" --ignore '**/hack/chglog/**' --rules /node_modules/sentences-per-line/index.js .

# base toolchain image
FROM ${TOOLCHAIN} AS toolchain
RUN apk --update --no-cache add bash curl build-base protoc protobuf-dev

# build tools
FROM toolchain AS tools
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOPATH /go
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b /bin v1.41.1
ARG GOFUMPT_VERSION
RUN go install mvdan.cc/gofumpt/gofumports@${GOFUMPT_VERSION} \
	&& mv /go/bin/gofumports /bin/gofumports

# tools and sources
FROM tools AS base
WORKDIR /src
COPY ./go.mod .
COPY ./go.sum .
RUN --mount=type=cache,target=/go/pkg go mod download
RUN --mount=type=cache,target=/go/pkg go mod verify
COPY ./cmd ./cmd
COPY ./pkg ./pkg
RUN --mount=type=cache,target=/go/pkg go list -mod=readonly all >/dev/null

# builds capi-darwin-amd64
FROM base AS capi-darwin-amd64-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=amd64 GOOS=darwin go build -ldflags "-s -w" -o /capi-darwin-amd64

# builds capi-darwin-arm64
FROM base AS capi-darwin-arm64-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=arm64 GOOS=darwin go build -ldflags "-s -w" -o /capi-darwin-arm64

# builds capi-linux-amd64
FROM base AS capi-linux-amd64-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o /capi-linux-amd64

# builds capi-linux-arm64
FROM base AS capi-linux-arm64-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=arm64 GOOS=linux go build -ldflags "-s -w" -o /capi-linux-arm64

# builds capi-linux-armv7
FROM base AS capi-linux-armv7-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=arm GOARM=7 GOOS=linux go build -ldflags "-s -w" -o /capi-linux-armv7

# builds capi-windows-amd64.exe
FROM base AS capi-windows-amd64.exe-build
COPY --from=generate / /
WORKDIR /src/cmd/capi
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOARCH=amd64 GOOS=windows go build -ldflags "-s -w" -o /capi-windows-amd64.exe

# runs gofumpt
FROM base AS lint-gofumpt
RUN find . -name '*.pb.go' | xargs -r rm
RUN find . -name '*.pb.gw.go' | xargs -r rm
RUN FILES="$(gofumports -l -local github.com/talos-systems/capi-utils .)" && test -z "${FILES}" || (echo -e "Source code is not formatted with 'gofumports -w -local github.com/talos-systems/capi-utils .':\n${FILES}"; exit 1)

# runs golangci-lint
FROM base AS lint-golangci-lint
COPY .golangci.yml .
ENV GOGC 50
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.cache/golangci-lint --mount=type=cache,target=/go/pkg golangci-lint run --config .golangci.yml

# runs unit-tests with race detector
FROM base AS unit-tests-race
ARG TESTPKGS
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg --mount=type=cache,target=/tmp CGO_ENABLED=1 go test -v -race -count 1 ${TESTPKGS}

# runs unit-tests
FROM base AS unit-tests-run
ARG TESTPKGS
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg --mount=type=cache,target=/tmp go test -v -covermode=atomic -coverprofile=coverage.txt -coverpkg=${TESTPKGS} -count 1 ${TESTPKGS}

FROM scratch AS capi-darwin-amd64
COPY --from=capi-darwin-amd64-build /capi-darwin-amd64 /capi-darwin-amd64

FROM scratch AS capi-darwin-arm64
COPY --from=capi-darwin-arm64-build /capi-darwin-arm64 /capi-darwin-arm64

FROM scratch AS capi-linux-amd64
COPY --from=capi-linux-amd64-build /capi-linux-amd64 /capi-linux-amd64

FROM scratch AS capi-linux-arm64
COPY --from=capi-linux-arm64-build /capi-linux-arm64 /capi-linux-arm64

FROM scratch AS capi-linux-armv7
COPY --from=capi-linux-armv7-build /capi-linux-armv7 /capi-linux-armv7

FROM scratch AS capi-windows-amd64.exe
COPY --from=capi-windows-amd64.exe-build /capi-windows-amd64.exe /capi-windows-amd64.exe

FROM scratch AS unit-tests
COPY --from=unit-tests-run /src/coverage.txt /coverage.txt

FROM capi-linux-${TARGETARCH} AS capi

FROM scratch AS image-capi
ARG TARGETARCH
COPY --from=capi capi-linux-${TARGETARCH} /capi
COPY --from=image-fhs / /
COPY --from=image-ca-certificates / /
LABEL org.opencontainers.image.source https://github.com/talos-systems/capi-utils
ENTRYPOINT ["/capi"]

