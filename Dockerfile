FROM golang:1.22-bookworm AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/api /app/api
COPY db/migrations /app/db/migrations
ENV HTTP_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/api"]
