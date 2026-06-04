FROM alpine:latest
WORKDIR /root/
COPY server .
COPY migrations ./migrations
COPY configs ./configs
COPY docs ./docs
EXPOSE 8080
CMD ["./server"]
