# Stage 1: Build the application
FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app

# Stage 2: Create the final runtime image
FROM busybox:1.33.1

WORKDIR /app

ENV SMS_PORT=8080
ENV SENDER_EMAIL_ADDR=""
ENV SENDER_EMAIL_PASS=""
ENV RECEIVER_EMAIL_ADDR=""

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
