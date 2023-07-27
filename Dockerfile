#build stage
FROM golang:1.21rc3 AS builder
WORKDIR /go/src/
COPY . .
RUN CGO_ENABLED=0 go build -o health-check-exporter

#final stage

FROM alpine:3.18.2
WORKDIR /app
COPY --from=builder /go/src/health-check-exporter ./health-check-exporter
ENTRYPOINT ["./health-check-exporter"]
EXPOSE 8080
