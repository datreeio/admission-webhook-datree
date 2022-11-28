FROM golang:1.18-alpine AS builder

ARG BUILD_ENVIRONMENT
ARG WEBHOOK_VERSION

WORKDIR /go/src/app



# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download

COPY . .
# cache the build
RUN --mount=type=cache,target=/root/.cache/go-build go build -o cert-generator -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" ./cmd/cert-generator

FROM alpine:3.14
COPY --from=builder /go/src/app/cert-generator /
EXPOSE 8443
ENTRYPOINT ["/cert-generator"]
