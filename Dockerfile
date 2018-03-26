FROM golang:1.9.4 AS builder
WORKDIR /go
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOPATH=/go
COPY ./main.go .
RUN go build -o grafana-ambassador main.go

FROM scratch
COPY --from=builder /go/grafana-ambassador /usr/bin/grafana-ambassador
ENTRYPOINT ["/usr/bin/grafana-ambassador"]
