FROM golang AS build
WORKDIR /go/src/github.com/qnib/jupyterport
COPY ./ ./
RUN go build

FROM debian
WORKDIR /app/
COPY --from=build /go/src/github.com/qnib/jupyterport/jupyterport ./jupyterport
COPY ./tpl ./tpl
ENV JUPYTERPORT_SPAWNER=docker \
    JUPYTERPORT_ADDR=:8080 \
    JUPYTERPORT_DEBUG=false
CMD ["/app/jupyterport"]
