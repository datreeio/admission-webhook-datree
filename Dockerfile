# Stage 1: Builder
FROM --platform=$BUILDPLATFORM golang:1.17-alpine AS builder

ARG BUILDPLATFORM
ARG BUILD_ENVIRONMENT
ARG WEBHOOK_VERSION

WORKDIR /go/src/app

# Copy and download dependencies (separately for caching)
COPY go.* ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -tags "$BUILD_ENVIRONMENT" \
    -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" \
    -o /go/bin/webhook-datree

# Stage 2: Final Image
FROM --platform=$TARGETPLATFORM alpine:3.14

# Copy the built binary from the previous stage
COPY --from=builder /go/bin/webhook-datree /usr/local/bin/

EXPOSE 8443

ENTRYPOINT ["webhook-datree"]
