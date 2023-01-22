FROM debian AS lib_builder

WORKDIR /foundry

RUN apt-get update -y && apt-get install -y \
  build-essential \
  cmake \
  git

RUN git clone https://github.com/jgarff/rpi_ws281x.git \
  && cd rpi_ws281x \
  && mkdir build \
  && cd build \
  && cmake -D BUILD_SHARED=OFF -D BUILD_TEST=OFF .. \
  && cmake --build . \
  && make install


FROM golang:1.19-bullseye as builder
COPY --from=lib_builder /usr/local/lib/libws2811.a /usr/local/lib/
COPY --from=lib_builder /usr/local/include/ws2811 /usr/local/include/ws2811
RUN mkdir -p /go/src/big-switch
COPY . /go/src/big-switch/
RUN cd /go/src/big-switch && make pi

FROM scratch
COPY --from=builder /go/src/big-switch/big-switch /