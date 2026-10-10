FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/enginer ./cmd/api

FROM alpine:3.20
COPY --from=build /out/enginer /usr/local/bin/enginer
EXPOSE 9091
ENTRYPOINT ["enginer"]
