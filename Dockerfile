FROM golang:1.24
WORKDIR /src
COPY . .
RUN go build -o /bin/app .
ENTRYPOINT ["/bin/app"]