FROM golang:1.18-alpine AS builder

ARG BUILD_ENVIRONMENT
ARG WEBHOOK_VERSION

WORKDIR /go/src/app

# download dependencies in a separate step to allow caching
COPY go.* .
RUN go mod download

COPY . .
# cache the build, 
## map the /root/.cache/go-build to your host go build cache folder
# RUN --mount=type=cache,target=/root/.cache/go-build go build ./cmd/webhook-datree -tags $BUILD_ENVIRONMENT -ldflags="-X github.com/datreeio/admission-webhook-datree/pkg/config.WebhookVersion=$WEBHOOK_VERSION" -o webhook-datree
RUN go build ./cmd/webhook-server

FROM alpine:3.14
COPY --from=builder /go/src/app/webhook-server /
EXPOSE 8443
ENTRYPOINT ["/webhook-server"]
