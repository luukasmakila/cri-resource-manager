FROM debian:stable-slim

COPY . .

RUN mv meme /bin

CMD ["meme", "-bs", "1G", "-bwc", "1", "-bws", "100M", "-bwi", "1s", "-bwod", "17M", "-brc", "0", "-p", "10s", "-ttl", "24h"]
