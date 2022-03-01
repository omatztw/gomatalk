FROM golang:1.17.7-buster as builder

RUN mkdir -p /workspace
WORKDIR /workspace

COPY .  /workspace/.

RUN apt update 
RUN apt install -y libopus-dev

RUN go build

FROM debian:buster-slim

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
RUN mkdir data
VOLUME /workspace/data
RUN mkdir wav
VOLUME /workspace/wav
RUN mkdir voices
VOLUME /workspace/voices

CMD ["/workspace/gomatalk", "-f", "/workspace/config/config.toml"]
