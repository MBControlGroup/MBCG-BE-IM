FROM golang:1.10.1
RUN mkdir -p /go/src/mbcs_im
ADD . /go/src/mbcs_im
WORKDIR /go/src/mbcs_im
CMD ["go", "run", "main.go"]
