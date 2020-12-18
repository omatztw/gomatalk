FROM golang:1.15.6-buster

RUN apt update 
RUN apt install -y build-essential curl unzip ffmpeg

RUN mkdir -p /workspace
WORKDIR /workspace

RUN \
    cd /usr/local/src/ && \
    curl -SLO http://downloads.sourceforge.net/hts-engine/hts_engine_API-1.10.tar.gz && \
    tar -zxvf hts_engine_API-1.10.tar.gz && \
    cd hts_engine_API-1.10 && \ 
    ./configure && make && \
    make install && \
    rm -rf hts_engine_API-1.10*

RUN \
     cd /usr/local/src/  && \
     curl -SLO http://downloads.sourceforge.net/open-jtalk/open_jtalk-1.11.tar.gz && \
     tar -zxvf open_jtalk-1.11.tar.gz && \ 
     cd open_jtalk-1.11 && \
     ./configure --with-hts-engine-header-path=/usr/local/include --with-hts-engine-library-path=/usr/local/lib && \
     make && \
     make install && \
     rm -rf open_jtalk-1.11*

RUN \
    mkdir -p /usr/share/open_jtalk/dic && \
    curl -SLO http://downloads.sourceforge.net/open-jtalk/open_jtalk_dic_utf_8-1.11.tar.gz && \
    tar -zxvf open_jtalk_dic_utf_8-1.11.tar.gz && \
    cp -r open_jtalk_dic_utf_8-1.11/* /usr/share/open_jtalk/dic && \
    rm -rf open_jtalk_dic_utf_8-1.11*

RUN \
 mkdir -p /usr/share/open_jtalk/voices && \
 curl -SLO http://downloads.sourceforge.net/open-jtalk/hts_voice_nitech_jp_atr503_m001-1.05.tar.gz && \
 tar -zxvf hts_voice_nitech_jp_atr503_m001-1.05.tar.gz && \
 cp hts_voice_nitech_jp_atr503_m001-1.05/nitech_jp_atr503_m001.htsvoice  /usr/share/open_jtalk/voices/. && \
 rm -rf hts_voice_nitech_jp_atr503_m001-1.05*

RUN \
 mkdir -p /usr/share/open_jtalk/voices && \
 curl -SLO https://downloads.sourceforge.net/project/mmdagent/MMDAgent_Example/MMDAgent_Example-1.8/MMDAgent_Example-1.8.zip && \
 unzip MMDAgent_Example-1.8.zip && \
 cp MMDAgent_Example-1.8/Voice/mei/*.htsvoice /usr/share/open_jtalk/voices/. && \
 rm -rf MMDAgent_Example-1.8*

COPY .  /workspace/.

RUN go build

VOLUME /workspace/data

CMD ["/workspace/gomatalk", "-f", "/workspace/config/config.toml"]