FROM debian:stable-slim

COPY . .

RUN mv meme /bin

CMD ["meme", "-bs", "500M", "-bwc", "1", "-bws", "100M", "-bwo", "200M", "-brc", "0", "-ttl", "24h"]
