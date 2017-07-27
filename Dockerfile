FROM golang:alpine
WORKDIR /go/src/github.com/3Blades/docker-events/

RUN apk update && apk upgrade && \
	apk add --no-cache git

COPY . .
RUN CGO_ENABLED=0 go get -t -v ./...
RUN CGO_ENABLED=0 go build -o events .

FROM scratch

EXPOSE 8000
CMD ["/events"]
COPY --from=0 /go/src/github.com/3Blades/docker-events/events .
