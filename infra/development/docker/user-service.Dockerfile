FROM alpine
WORKDIR /app

COPY build build

ENTRYPOINT build/user-service
