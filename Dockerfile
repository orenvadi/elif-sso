ARG APP_NAME=myapp

# Build stage
FROM golang:1.22 as build
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /myapp
COPY . .
RUN go mod download
RUN go build -o /app/$APP_NAME  

# Production stage
FROM alpine:latest as production
ARG APP_NAME
ENV APP_NAME=$APP_NAME
WORKDIR /root/
COPY --from=build /myapp/$APP_NAME ./  

# Copy migration files
COPY migrations /migrations

# Run migrator
RUN ./migrator

# Command to run the application
CMD ./$APP_NAME
