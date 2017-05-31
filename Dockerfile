FROM scratch

EXPOSE 8000
CMD ["/stats"]
ADD stats /
