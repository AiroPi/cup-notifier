FROM golang:1.26.3 AS build

WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cup-notifier

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /cup-notifier /cup-notifier
CMD ["/cup-notifier"]

