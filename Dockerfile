FROM golang:1.24.1-alpine3.20 AS base

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

    
FROM base as builder

WORKDIR /app

COPY . .

# RUN go mod download

RUN go build -o backend .


FROM golang:1.24.1-alpine3.20 AS RUNNER

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

WORKDIR /home/node

COPY --from=builder /app/backend ./backend

EXPOSE $APP_PORT

CMD [ "./backend" ]