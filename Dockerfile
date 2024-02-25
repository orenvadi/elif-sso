ARG APP_NAME=app

# Build stage
FROM golang:1.22 as build
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /$APP_NAME

# Production stage
FROM alpine:latest as production
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /root/
COPY --from=build /$APP_NAME ./

# Copy migration files
COPY migrations /migrations

# Run migrator
RUN ./migrator

# Command to run the application
CMD ./$APP_NAME
