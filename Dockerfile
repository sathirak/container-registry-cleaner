FROM golang:1.24 AS builder
WORKDIR /src
COPY . .
RUN go build -o /bin/app .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /bin/app /bin/app
ENTRYPOINT ["/bin/app"]