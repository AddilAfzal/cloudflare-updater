FROM golang:1.15.2-alpine3.12

WORKDIR $GOPATH/src/github.com/AddilAfzal/cloudflare-updater

COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

RUN go build -o bin

RUN chmod +x bin && \
    mkdir /app/ && \
    mv bin /app/bin

# Run the executable
ENTRYPOINT ["/app/bin"]
