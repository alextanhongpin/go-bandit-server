FROM golang:1.10.2 as builder

WORKDIR /go/src

# Copy vendor files
COPY vendor/ .

WORKDIR /go/src/github.com/alextanhongpin/go-bandit-server

COPY main.go model.go store.go schema.go go.mod ./

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app/
COPY --from=builder /go/src/github.com/alextanhongpin/go-bandit-server/app .

# Metadata params
ARG VERSION
ARG BUILD_DATE
ARG VCS_URL
ARG VCS_REF
ARG NAME
ARG VENDOR

ENV VCS_REF=${VCS_REF}
ENV VERSION=${VERSION}

# Metadata
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name=$NAME \
      org.label-schema.description="Example of multi-stage docker build" \
      org.label-schema.url="https://example.com" \
      org.label-schema.vcs-url=https://github.com/alextanhongpin/$VCS_URL \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vendor=$VENDOR \
      org.label-schema.version=$VERSION \
      org.label-schema.docker.schema-version="1.0" \
      org.label-schema.docker.cmd="docker run -d alextanhongpin/hello-world"

CMD ["./app"]
