############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/gserve
COPY src/gserve .
COPY html /html
# Fetch dependencies.
RUN go mod init
# Using go get.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/gserve
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /html /html
COPY --from=builder /go/bin/gserve /go/bin/gserve
# Run the hello binary.
ENTRYPOINT ["/go/bin/gserve"]