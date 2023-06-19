FROM debian:stable-slim

RUN apt-get update && apt-get install -y software-properties-common build-essential gnupg make curl sudo git socat vim procps

# Install Go 1.19
RUN curl -LO https://golang.org/dl/go1.19.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz && \
    rm go1.19.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

COPY . .

RUN GO111MODULE=on go get -u github.com/containerd/nri/pkg/api
RUN GO111MODULE=on go get -u github.com/containerd/nri/pkg/stub
RUN GO111MODULE=on go get -u github.com/sirupsen/logrus
RUN GO111MODULE=on go get -u gopkg.in/yaml.v2

# Build the plugin binary
RUN go build -gcflags "all=-N -l"

# Make the plugin directory & add the plugin there
RUN sudo mkdir -p /opt/nri/plugins/
RUN mv memtier-nri /opt/nri/plugins/10-memtier-nri

RUN mv memtierd /bin

CMD ["/opt/nri/plugins/10-memtier-nri"]
