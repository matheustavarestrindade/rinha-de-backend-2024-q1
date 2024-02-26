FROM golang:1.22

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY . .

# RUN go build -race -o /usr/local/bin/app .
RUN go build -o /usr/local/bin/app .

ENTRYPOINT ["app"]
