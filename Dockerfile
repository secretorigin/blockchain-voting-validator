FROM golang:1.19-alpine

WORKDIR /app

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod tidy

RUN go build -o /service cmd/main.go

CMD /service