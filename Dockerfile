FROM golang:alpine AS DOWNLOAD

WORKDIR /app
COPY ./download_samples.sh /app/download_samples.sh
RUN sh /app/download_samples.sh

FROM golang:alpine

WORKDIR /app
COPY ./ /app
ENV GO111MODULE=on

COPY --from=DOWNLOAD /app/samples /app/samples
RUN apk update && apk add postgresql-client
RUN go install github.com/lukasmartinelli/pgfutter

CMD ["/bin/sh", "/app/test.sh"]