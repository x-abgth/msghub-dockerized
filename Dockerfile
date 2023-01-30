# Building stage
FROM golang:1.19.3-alpine AS builder

# maintainer info
LABEL maintainer = "Abhijith A <abhijithak683@gmail.com>"

WORKDIR /app

COPY ./msghub-server/go.mod ./

RUN go mod download

COPY . .

RUN cd msghub-server && go build -o main


# Final running stage
FROM alpine:latest

WORKDIR /app

RUN mkdir msghub-client

COPY --from=builder /app/msghub-client/ ./msghub-client/

COPY ./.env .

COPY --from=builder /app/msghub-server/main .

RUN chmod +x ./main

CMD ["./main"]

EXPOSE 9000

