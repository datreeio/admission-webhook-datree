FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS builder
ARG BUILD_ENVIRONMENT
ARG TARGETARCH
ARG TARGETOS
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG WEBHOOK_VERSION

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
