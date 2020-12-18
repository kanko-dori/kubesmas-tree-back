FROM golang:1.15 AS builder
ADD . /work
WORKDIR /work
RUN go build -o ./app

#

FROM scratch

COPY --from=builder /work/app /usr/local/bin/app

ENTRYPOINT ["/usr/local/bin/app"]