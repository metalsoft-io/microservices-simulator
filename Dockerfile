# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/MetalSoft-io/microservices-simulator

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/microservices-simulator

# Document that the service listens on port 3365.
EXPOSE 3365
