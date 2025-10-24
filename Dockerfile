FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod .
COPY . .
RUN go build -o server ./

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
