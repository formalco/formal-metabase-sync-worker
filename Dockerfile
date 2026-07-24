FROM --platform=$BUILDPLATFORM golang AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

COPY . .

RUN go mod tidy && \
  go mod download && \
  go mod verify && \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags='-w -s -extldflags "-static"' -o worker .

############################
# STEP 2 build a small image
############################
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/worker /app/worker

CMD ["./app/worker"]
