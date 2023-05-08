FROM golang:1.20.3-alpine3.16

RUN mkdir -p home/app

COPY . /home/app

WORKDIR /home/app

RUN apk add --no-cache graphviz

#CMD go tool pprof -svg -alloc_objects privatapi mem_out > /home/app/data/mem_ao.svg
CMD go tool pprof -svg privatapi cpu_out > /home/app/data/cpu.svg