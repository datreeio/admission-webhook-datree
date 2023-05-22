FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS builder
ARG BUILD_ENVIRONMENT
ARG TARGETARCH
ARG TARGETOS
ARG TARGETPLATFORM
ARG TARGETVARIANT
ARG BUILDPLATFORM
ARG TARGETPLATFORM

ARG WEBHOOK_VERSION
RUN echo "Building for $BUILD_ENVIRONMENT"
RUN echo "Webhook version: $WEBHOOK_VERSION"
RUN echo "Target OS: $TARGETOS"
RUN echo "Target ARCH: $TARGETARCH"
RUN echo "Target PLATFORM: $TARGETPLATFORM"
RUN echo "Target VARIANT: $TARGETVARIANT"
RUN echo "Build PLATFORM: $BUILDPLATFORM"
RUN echo "Build PLATFORM: $TARGETPLATFORM"


WORKDIR /go/src/app
# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download
COPY . .
# cache the build
RUN --mount=type=cache,target=/root/.cache/go-build GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" -o webhook-datree
FROM --platform=$TARGETPLATFORM alpine:3.14
COPY --from=builder /go/src/app/webhook-datree /
EXPOSE 8443
ENTRYPOINT ["/webhook-datree"]