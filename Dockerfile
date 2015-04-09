FROM google/golang

RUN mkdir -p /gopath/src/github.com/lavab/invite-api
ADD . /gopath/src/github.com/lavab/invite-api
RUN go get github.com/lavab/invite-api

CMD []
ENTRYPOINT ["/gopath/bin/invite-api"]
