FROM golang AS build
WORKDIR /go/src/github.com/qnib/jupyterport
COPY ./ ./
RUN go build

FROM debian
RUN apt-get update \
 && apt-get install -y curl \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app/
COPY ./tpl ./tpl
COPY --from=build /go/src/github.com/qnib/jupyterport/jupyterport ./jupyterport
ENV JUPYTERPORT_SPAWNER=docker \
    JUPYTERPORT_ADDR=:8080 \
    JUPYTERPORT_DEBUG=false
CMD ["/app/jupyterport"]
