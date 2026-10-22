FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/hub-api ./apps/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/hub-api /bin/hub-api
EXPOSE 8090
ENTRYPOINT ["/bin/hub-api"]
