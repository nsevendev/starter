FROM golang:1.24.4
WORKDIR /app

RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/spf13/cobra-cli@latest

CMD ["CompileDaemon", "--build=go build -o tmp/starter .", "--command=tmp/starter"]