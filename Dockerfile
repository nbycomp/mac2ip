FROM golang:1.12.3 as base

RUN mkdir -p /opt/app
WORKDIR /opt/app

COPY . ./

RUN go mod vendor

FROM base as builder

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app .

FROM scratch

COPY --from=builder /app ./

ENTRYPOINT ["./app"]
