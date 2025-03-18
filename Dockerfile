FROM golang:1.24-alpine
WORKDIR /opt/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o vocabforgebin ./cmd/bot/main.go

FROM alpine:3.21
EXPOSE 8080
COPY --from=0 /opt/app/vocabforgebin /bin/vocabforgebin
CMD ["/bin/vocabforgebin"]
