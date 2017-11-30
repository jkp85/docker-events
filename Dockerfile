FROM alpine
RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*

EXPOSE 8000
CMD ["/events"]
ADD events /
