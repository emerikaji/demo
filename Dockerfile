FROM --platform=$BUILDPLATFORM golang:1.27 AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o main .

FROM scratch
USER 65534:65534

COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]