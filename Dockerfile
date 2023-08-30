# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.19 as builder

# https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
ENV CGO_ENABLED=0

# Set the current working directory inside the container.
WORKDIR /go/src/github.com/score-spec/score-humanitec

# Copy the entire project and build it.
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/score-humanitec ./cmd/score-humanitec

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3

# Set the current working directory inside the container.
WORKDIR /score-humanitec

# Copy the binary from the builder image.
COPY --from=builder /usr/local/bin/score-humanitec /usr/local/bin/score-humanitec

# Run the binary.
ENTRYPOINT ["/usr/local/bin/score-humanitec"]

FROM builder AS environment

RUN export PATH=$PATH:/usr/local/bin/score-humanitec
