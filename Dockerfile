FROM scratch

EXPOSE 8000
CMD ["/events"]
ADD events /
