###########
# Stage 1 #
###########
FROM golang:1.11-alpine as build_base

RUN apk add git

ENV GO111MODULE=on

WORKDIR /scrimplb
COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

###########
# Stage 2 #
###########
FROM build_base as build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o scrimplb .

###########
# Stage 3 #
###########
FROM alpine as image

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=build /scrimplb/scrimplb /scrimplb

EXPOSE 9999

ENTRYPOINT ["/scrimplb"]
