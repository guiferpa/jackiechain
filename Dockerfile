FROM golang:1.19 AS builder

WORKDIR /opt/jackie

COPY . /opt/jackie

RUN make prod

FROM scratch

COPY --from=builder /opt/jackie/jackie /usr/bin/jackie

CMD ["jackie"]
