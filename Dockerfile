FROM debian:stable-slim

RUN apt-get update && apt-get install -y wget npm python3 chromium xvfb libxss1 libpango1.0-0 libnss3 libx11-xcb1 libgbm-dev
RUN wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
RUN rm go1.22.5.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /app
COPY . .
RUN chmod +x start.sh
RUN go build -o celeve

WORKDIR /app/web
RUN ls
RUN npm install
RUN npm run build

EXPOSE 8989
EXPOSE 23538
CMD ["/app/start.sh"]
