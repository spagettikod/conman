#
#   docker build -t spagettikod/conman . && docker run --rm -p 26652:80  -v /home/roland/development/conman/www:/www -v /var/run/docker.sock:/var/run/docker.sock spagettikod/conman
#

FROM golang:1.15 AS golang
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"'

FROM scratch
COPY --from=golang /go/src/app/conman /conman
COPY --from=golang /go/src/app/www /www
ENTRYPOINT [ "/conman" ]