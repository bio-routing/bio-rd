FROM golang as builder
ADD . /go/bio
WORKDIR /go/bio/cmd/bio-rd
RUN GOOS=linux go build -o /go/bin/bio-rd

FROM debian:stable
WORKDIR /app
ENV CMD_ARGS="--config.file=/config/bio-rd.yml"
COPY --from=builder /go/bin/bio-rd .
CMD /app/bio-rd ${CMD_ARGS}
VOLUME /config
EXPOSE 179
EXPOSE 5566
EXPOSE 55667
