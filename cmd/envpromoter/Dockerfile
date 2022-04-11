FROM golang:1.17.5-alpine3.15 as builder

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

# moves for make compilation happy
RUN GO111MODULE=on CGO_ENABLED=0 go build --ldflags "-s -w" -o app .



# FROM gcr.io/distroless/static
FROM alpine:3.12

ARG DIR

WORKDIR /usr/local/bin/

# Add executable into image
COPY --from=builder  /app  /usr/local/bin/


ENTRYPOINT [ "./app" ]