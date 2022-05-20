FROM golang:1.18-bullseye AS builder

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN CGO_ENABLED=1 GOOS=linux go build -o app main.go

FROM debian:bullseye-slim AS production

COPY --from=builder /app .

CMD ["./app"]
