FROM golang:1.23-alpine as base

WORKDIR /work

RUN addgroup -S homework && adduser -S homework -G homework
RUN chown homework:homework /work

RUN apk update && apk add git
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz \
    | tar xvz && mv migrate /usr/local/bin/migrate


RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY go.mod go.mod
COPY go.sum go.sum
COPY example.env .env

RUN go mod download

FROM base AS dev

VOLUME [ "./cmd" ]
VOLUME [ "./internal" ]
VOLUME [ "./migrations" ]


CMD ["go", "run", "./cmd/server/main.go"]


FROM base as build

COPY cmd/ cmd/
COPY internal/ internal/
COPY migrations/ migrations/
COPY .env .env

RUN go build -o /tmp/server ./cmd/server/main.go

USER homework

FROM scratch as deploy

WORKDIR /work

COPY --from=build /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=build /tmp/server /usr/local/bin/server

CMD ["/usr/local/bin/server"]
