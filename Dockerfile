FROM golang:1.23.4-alpine3.20 AS build

WORKDIR /opt

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o pongo cmd/server/pongo.go

FROM alpine:3.20 AS runtime

WORKDIR /opt

COPY --from=build /opt/.env .env
COPY --from=build /opt/pongo pongo

EXPOSE 8080

CMD ["./pongo"]
