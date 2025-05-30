FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy 

COPY . .

RUN go build -o main cmd/server/main.go

EXPOSE 3000

CMD ["./main"]