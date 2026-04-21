FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker

FROM python:3.12-alpine
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/worker /app/worker
ENTRYPOINT ["/app/worker"]
