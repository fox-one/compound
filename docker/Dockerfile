FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN make build-prod

FROM frolvlad/alpine-glibc:latest

WORKDIR /app

COPY --from=builder /app/builds/rings .

ENTRYPOINT ["/app/rings"]
