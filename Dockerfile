FROM golang:1.14 as builder

RUN  apt-get update && apt-get -y install bash git make zip bzr && apt-get clean && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*
ADD . /go/src/github.com/llomgui/terraform-provider-openshift
WORKDIR /go/src/github.com/llomgui/terraform-provider-openshift
ENV GOPROXY=https://proxy.golang.org
RUN ["make", "tools", "build"]

###

FROM hashicorp/terraform:0.12.23

COPY --from=builder /go/src/github.com/llomgui/terraform-provider-openshift/bin/* /bin/

VOLUME ["/workdir"]
WORKDIR /workdir

ENTRYPOINT ["/bin/terraform"]
CMD ["--help"]
