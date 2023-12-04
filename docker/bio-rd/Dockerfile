FROM golang as builder
ADD . /go/bio-rd
WORKDIR /go/bio-rd/cmd/bio-rd
RUN GOOS=linux go build -o /go/bin/bio-rd

FROM debian:stable
WORKDIR /app
COPY --from=builder /go/bin/bio-rd .
CMD /app/bio-rd --config.file=/config/bio-rd.yml ${CMD_ARGS}
#RUN apt update && \
#    apt install -y libcap2-bin && \
#    setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip /app/bio-rd && \
#    useradd --system bio-rd
#USER bio-rd
VOLUME /config
EXPOSE 179
EXPOSE 5566
EXPOSE 55667
