# We'll choose the incredibly lightweight
# Go alpine image to work with
FROM docker.io/golang:1.18-alpine3.14 AS builder

# We create an /app directory in which
# we'll put all of our project code
RUN mkdir /app
WORKDIR /app
RUN apk update
RUN apk add gcc libstdc++ libc-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
# We want to build our application's binary executable
RUN go build -o main -a ./cmd/web/main.go


# the lightweight scratch image we'll
# run our application within
FROM docker.io/alpine:3.14 AS production
# We have to copy the output from our
# builder stage to our production stage
COPY --from=builder /app/main .

RUN apk update
RUN apk upgrade
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
RUN echo "America/Los_Angeles" >  /etc/timezone
RUN apk del tzdata

VOLUME /config
EXPOSE 2112


CMD ["./main"]
