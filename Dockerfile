FROM golang:1.19-alpine AS builder
ARG BUILD_ENVIRONMENT
ARG TARGETARCH

ARG WEBHOOK_VERSION
RUN echo "Building for $BUILD_ENVIRONMENT"
RUN echo "Webhook version: $WEBHOOK_VERSION"
RUN echo "Target OS: $TARGETOS"
RUN echo "Target ARCH: $TARGETARCH"
RUN echo "Target PLATFORM: $TARGETPLATFORM"

WORKDIR /go/src/app
# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download
COPY . .
# cache the build
RUN --mount=type=cache,target=/root/.cache/go-build GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" -o webhook-datree
FROM --platform=$TARGETARCH alpine:3.14
COPY --from=builder /go/src/app/webhook-datree /
EXPOSE 8443
ENTRYPOINT ["/webhook-datree"]