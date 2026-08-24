# Multi-arch (amd64 + arm64) build of Prow's peribolos, for the NeoNephos
# account-administration automation.
#
# Upstream publishes only linux/amd64 (built with ko), which forces the
# reconcile CI job onto amd64 runners. We rebuild the same pinned source for
# both architectures and publish to ghcr.io/neonephos/account-administration/peribolos
# so the job can run on ARM runners too.
#
# The binary is placed at /ko-app/peribolos to match the upstream image path,
# so the Taskfile invocation is unchanged.
#
# PROW_VERSION must match the pinned commit (see build-peribolos.yml). The image
# tag we mirror is vYYYYMMDD-<shortsha>; PROW_REF is the full commit sha.

ARG GO_VERSION=1.26

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS build
ARG TARGETOS
ARG TARGETARCH
# Full commit sha of kubernetes-sigs/prow to build from.
ARG PROW_REF
# Human-readable version stamped into the binary (e.g. v20260821-a61940897).
ARG PROW_VERSION=dev

WORKDIR /src
# Fetch the pinned prow source. A shallow clone of the exact revision keeps this
# fast and reproducible.
RUN git init -q . \
    && git remote add origin https://github.com/kubernetes-sigs/prow.git \
    && git fetch -q --depth 1 origin "${PROW_REF}" \
    && git checkout -q FETCH_HEAD

# Cross-compile peribolos for the target platform. CGO is off so the binary is
# static and runs on distroless.
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags "-s -w -X sigs.k8s.io/prow/pkg/version.Version=${PROW_VERSION} -X sigs.k8s.io/prow/pkg/version.Name=peribolos" \
      -o /out/peribolos \
      ./cmd/peribolos

# Alpine base: matches upstream (which uses a custom alpine) and, crucially,
# provides /bin/sh. The Taskfile writes the token to a file via
# `--entrypoint sh -c '... > /tmp/token; exec /ko-app/peribolos ...'`, which
# needs a shell in the image (distroless has none).
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /out/peribolos /ko-app/peribolos
USER 65532:65532
ENTRYPOINT ["/ko-app/peribolos"]
