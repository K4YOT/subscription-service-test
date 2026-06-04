FROM alpine:latest

WORKDIR /root/

# Копируем бинарник, миграции, конфиги и сгенерированную документацию
COPY server .
COPY migrations ./migrations
COPY configs ./configs
COPY docs ./docs

EXPOSE 8080

CMD ["./server"]