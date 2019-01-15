###########
# Stage 1 #
###########
FROM golang:1.11-alpine as build

RUN apk add git

WORKDIR /scrimplb
COPY . /scrimplb

RUN CGO_ENABLED=0 GOOS=linux go build -a -o scrimplb .

###########
# Stage 2 #
###########
FROM alpine as image

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=build /scrimplb/scrimplb /scrimplb

EXPOSE 9999

ENTRYPOINT ["/scrimplb"]
