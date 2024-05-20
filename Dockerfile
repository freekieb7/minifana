FROM golang:1.22-alpine

#RUN apk update && apk add --no-cache build-base cmake git

WORKDIR /app

COPY . .

RUN go install github.com/cosmtrek/air@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go mod download

CMD ["air", "-c", ".air.toml"]