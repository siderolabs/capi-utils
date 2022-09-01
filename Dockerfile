# syntax = docker/dockerfile-upstream:1.2.0-labs

# THIS FILE WAS AUTOMATICALLY GENERATED, PLEASE DO NOT EDIT.
#
# Generated on 2022-09-01T17:46:14Z by kres ec2cd64.

ARG TOOLCHAIN

# cleaned up specs and compiled versions
FROM scratch AS generate

FROM ghcr.io/siderolabs/ca-certificates:v1.1.0 AS image-ca-certificates

FROM ghcr.io/siderolabs/fhs:v1.1.0 AS image-fhs

# runs markdownlint
FROM docker.io/node:18.7.0-alpine3.16 AS lint-markdown
WORKDIR /src
RUN npm i -g markdownlint-cli@0.31.1
RUN npm i sentences-per-line@0.2.1
COPY .markdownlint.json .
COPY ./README.md ./README.md
RUN markdownlint --ignore "CHANGELOG.md" --ignore "**/node_modules/**" --ignore '**/hack/chglog/**' --rules node_modules/sentences-per-line/index.js .

# base toolchain image
FROM ${TOOLCHAIN} AS toolchain
RUN apk --update --no-cache add bash curl build-base protoc protobuf-dev

# build tools
FROM toolchain AS tools
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOPATH /go
ARG GOLANGCILINT_VERSION
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCILINT_VERSION} \
	&& mv /go/bin/golangci-lint /bin/golangci-lint
ARG GOFUMPT_VERSION
RUN go install mvdan.cc/gofumpt@${GOFUMPT_VERSION} \
	&& mv /go/bin/gofumpt /bin/gofumpt
ARG GOIMPORTS_VERSION
RUN go install golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION} \
	&& mv /go/bin/goimports /bin/goimports
ARG DEEPCOPY_VERSION
RUN go install github.com/siderolabs/deep-copy@${DEEPCOPY_VERSION} \
	&& mv /go/bin/deep-copy /bin/deep-copy

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
RUN FILES="$(gofumpt -l .)" && test -z "${FILES}" || (echo -e "Source code is not formatted with 'gofumpt -w .':\n${FILES}"; exit 1)

# runs goimports
FROM base AS lint-goimports
RUN FILES="$(goimports -l -local github.com/siderolabs/capi-utils .)" && test -z "${FILES}" || (echo -e "Source code is not formatted with 'goimports -w -local github.com/siderolabs/capi-utils .':\n${FILES}"; exit 1)

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
LABEL org.opencontainers.image.source https://github.com/siderolabs/capi-utils
ENTRYPOINT ["/capi"]

