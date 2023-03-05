FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

ARG BUILD_ENVIRONMENT
ARG WEBHOOK_VERSION

WORKDIR /go/src/app

# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download

COPY . .
# cache the build
RUN --target=$TARGETPLATFORM --mount=type=cache,target=/root/.cache/go-build go build -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" -o webhook-datree

FROM --platform=$BUILDPLATFORM alpine:3.14
COPY --from=builder /go/src/app/webhook-datree /
EXPOSE 8443
ENTRYPOINT ["/webhook-datree"]
