FROM golang:1.18-alpine AS builder

ARG BUILD_ENVIRONMENT
ARG WEBHOOK_VERSION

WORKDIR /go/src/app

# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download

COPY . .
# map /root/.cache/go-build to host go build cache folder
RUN --mount=type=cache,target=/root/.cache/go-build go build -o webhook-server -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" ./cmd/webhook-server 

FROM alpine:3.14
COPY --from=builder /go/src/app/webhook-server /
EXPOSE 8443
ENTRYPOINT ["/webhook-server"]
