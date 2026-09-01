FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY GoAWD/ .
RUN go build -o /goawd-server ./cmd/server && \
    go build -o /goawd-roundworm ./cmd/roundworm && \
    go build -o /goawd-guardian ./cmd/guardian

FROM node:18-alpine AS frontend
COPY Frontend/ /src/Frontend
WORKDIR /src/Frontend
RUN npm install --ignore-scripts --legacy-peer-deps && \
    npm run build

FROM alpine:latest
LABEL maintainer="AoiAWD Project"
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /goawd-server /usr/local/bin/goawd-server
COPY --from=builder /goawd-roundworm /usr/local/bin/goawd-roundworm
COPY --from=builder /goawd-guardian /usr/local/bin/goawd-guardian
COPY --from=frontend /src/Frontend/dist/ ./public/
EXPOSE 1337 8023
ENTRYPOINT ["goawd-server"]
CMD ["-storage", "file", "-file-path", "/data/goawd.json", "-public", "./public"]