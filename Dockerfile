# Stage 1: Build binary
FROM --platform=$BUILDPLATFORM golang:1.24-bullseye AS builder

WORKDIR /workspace

# Cache dependency layer
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/

# Compile static binary for target architecture
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=v0.1.0
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags="-s -w -X k8s.io/component-base/version.gitCommit=${GIT_COMMIT}" \
    -o /bin/tiered-node-scheduler ./cmd/scheduler

# Stage 2: Minimal non-root runtime image
FROM gcr.io/distroless/static:nonroot

USER 10001:10001
WORKDIR /

COPY --from=builder /bin/tiered-node-scheduler /usr/local/bin/tiered-node-scheduler

ENTRYPOINT ["/usr/local/bin/tiered-node-scheduler"]
