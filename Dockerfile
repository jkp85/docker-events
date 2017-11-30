FROM alpine AS base
RUN apk update && apk add ca-certificates
FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000
CMD ["/events"]
ADD events /
