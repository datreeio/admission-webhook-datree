FROM golang:1.18-alpine AS builder

ARG BUILD_ENVIRONMENT=staging
ARG WEBHOOK_VERSION="unknown"

WORKDIR /go/src/app

# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download

COPY . .
# cache the build
RUN --mount=type=cache,target=/root/.cache/go-build go build -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" -o webhook-datree

FROM alpine:3.14
COPY --from=builder /go/src/app/webhook-datree /
EXPOSE 8443
ENTRYPOINT ["/webhook-datree"]
