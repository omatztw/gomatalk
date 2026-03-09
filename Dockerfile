FROM golang:1.25.4 as builder

RUN mkdir -p /workspace
WORKDIR /workspace

COPY .  /workspace/.

RUN apt update 
RUN apt install -y libopus-dev wget unzip pkg-config

RUN \
 wget -O /tmp/libdave.zip https://github.com/discord/libdave/releases/download/v1.1.1/cpp/libdave-Linux-X64-boringssl.zip && \
 unzip /tmp/libdave.zip -d /tmp/libdave && \
 cp /tmp/libdave/lib/libdave.so /usr/local/lib/libdave.so && \
 cp /tmp/libdave/include/dave/dave.h /usr/local/include/dave.h && \
 mkdir -p /usr/local/lib/pkgconfig && \
 printf "prefix=/usr/local\nlibdir=\${prefix}/lib\nincludedir=\${prefix}/include\n\nName: dave\nDescription: Discord Audio/Video Encryption library\nVersion: 1.1.1\nLibs: -L\${libdir} -ldave\nCflags: -I\${includedir}\n" > /usr/local/lib/pkgconfig/dave.pc && \
 rm -rf /tmp/libdave /tmp/libdave.zip

ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

RUN go build

FROM debian:trixie-slim

RUN apt-get update 
RUN apt-get install -y build-essential unzip ffmpeg wget open-jtalk open-jtalk-mecab-naist-jdic

RUN mkdir -p /workspace
WORKDIR /workspace

RUN \
 mkdir -p /usr/share/open_jtalk/voices && \
 wget http://downloads.sourceforge.net/open-jtalk/hts_voice_nitech_jp_atr503_m001-1.05.tar.gz && \
 tar -zxvf hts_voice_nitech_jp_atr503_m001-1.05.tar.gz && \
 cp hts_voice_nitech_jp_atr503_m001-1.05/*  /usr/share/open_jtalk/voices/. && \
 rm -rf hts_voice_nitech_jp_atr503_m001-1.05*

RUN \
 mkdir -p /usr/share/open_jtalk/voices && \
 wget https://downloads.sourceforge.net/project/mmdagent/MMDAgent_Example/MMDAgent_Example-1.8/MMDAgent_Example-1.8.zip && \
 unzip MMDAgent_Example-1.8.zip && \
 cp MMDAgent_Example-1.8/Voice/mei/* /usr/share/open_jtalk/voices/. && \
 rm -rf MMDAgent_Example-1.8*


RUN apt-get purge -y --auto-remove build-essential wget unzip
RUN apt-get clean autoclean
RUN apt-get autoremove --yes
RUN rm -rf /var/lib/{apt,dpkg,cache,log}/


COPY --from=builder /workspace/gomatalk .
COPY --from=builder /usr/local/lib/libdave.so /usr/local/lib/libdave.so
ENV LD_LIBRARY_PATH=/usr/local/lib
RUN mkdir data
VOLUME /workspace/data
RUN mkdir wav
VOLUME /workspace/wav
RUN mkdir voices
VOLUME /workspace/voices
RUN mkdir migrates
COPY migrates  /workspace/migrates/.

CMD ["/workspace/gomatalk", "-f", "/workspace/config/config.toml"]
