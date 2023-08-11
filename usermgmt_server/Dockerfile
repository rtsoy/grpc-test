FROM golang:alpine as builder

WORKDIR /build

COPY . .

RUN go build -o main ./usermgmt_server/usermgmt_server.go

FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser

COPY --from=builder /build/main /app/

WORKDIR /app

EXPOSE 3000

CMD ["./main"]