FROM golang:1.16 as builder

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
COPY api api
COPY handler handler
COPY cmd cmd
RUN go build -o /go/bin/service -ldflags '-s -w' cmd/service/main.go
RUN go build -o /go/bin/gateway -ldflags '-s -w' cmd/gateway/main.go


FROM scratch as runner

COPY certs /certs
COPY --from=builder /go/bin/service /service
COPY --from=builder /go/bin/gateway /gateway

CMD ["/service"]
