FROM golang:1.17-alpine
WORKDIR /app
EXPOSE 3030
COPY . .
RUN go build -o /bin/gorest github.com/covenroven/gorest
CMD ["/bin/gorest"]
