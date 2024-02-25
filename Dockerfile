# Build stage
FROM golang:1.22 as build
ARG APP_NAME=app
ENV APP_NAME=$APP_NAME
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /$APP_NAME cmd/sso/main.go

# Production stage
FROM alpine:latest as production
ARG APP_NAME=app
ENV APP_NAME=$APP_NAME
WORKDIR /root/
COPY --from=build /$APP_NAME ./
CMD ./$APP_NAME
