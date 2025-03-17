FROM golang:1.24-alpine
WORKDIR /opt/app
COPY . .
RUN go mod download
RUN go build -o vocabforgebin

FROM alpine:3.21
COPY --from=0 /opt/app/vocabforgebin /bin/vocabforgebin
CMD ["/bin/vocabforgebin"]
