# Use the official Go image (includes Go toolchain)
FROM golang:1.24 AS builder
WORKDIR /app

# Only copy go.mod & go.sum to cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build the binary
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o lbipam-annotate main.go

# 2) Final stage
FROM alpine:3.18
# Install certs if your job ever needs HTTPS; skip if not needed
RUN apk add --no-cache ca-certificates

WORKDIR /app
# Copy the statically-built binary
COPY --from=builder /app/lbipam-annotate ./

# Run it
ENTRYPOINT ["./lbipam-annotate"]
