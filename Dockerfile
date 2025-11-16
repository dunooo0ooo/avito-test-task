FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrator ./cmd/migrator

FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/app /app/app
COPY --from=build /app/migrator /app/migrator
COPY --from=build /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/app"]