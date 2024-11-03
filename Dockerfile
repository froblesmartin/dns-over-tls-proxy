FROM golang:alpine3.19@sha256:282ddcdde8fea3bceb235a539d7a2fee2f52559c5319c42e577df4bcae2c2f39 as build

WORKDIR /app
COPY . ./
RUN go build -o /dot-proxy

FROM gcr.io/distroless/static-debian12:nonroot@sha256:e9ac71e2b8e279a8372741b7a0293afda17650d926900233ec3a7b2b7c22a246

COPY --from=build /dot-proxy /dot-proxy
CMD [ "/dot-proxy" ]
