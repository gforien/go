FROM golang:1.23-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o hello .

FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/hello .
CMD ["/app/hello", "--http", "--listen"]
