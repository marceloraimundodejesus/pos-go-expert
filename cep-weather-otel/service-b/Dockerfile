# build
FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# run
FROM gcr.io/distroless/base-debian12
ENV PORT=8080
EXPOSE 8080
WORKDIR /app
COPY --from=build /app/server /app/server
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
