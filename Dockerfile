FROM golang:1.22.5-alpine as BUILDER

WORKDIR /app-build

COPY ./ ./

RUN go build -o ContainerMonitor

FROM scratch

WORKDIR /app

COPY --from=BUILDER /app-build/src/ContainerMonitor ./

VOLUME /var/run/docker.sock

EXPOSE "9099"

ENTRYPOINT ["/app/ContainerMonitor"]

CMD ["-port=9099"]


