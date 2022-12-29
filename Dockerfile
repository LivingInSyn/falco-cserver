from golang:1.18

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# move everything into usr src app
COPY . .
RUN go build -v -o sfc ./...
RUN ln -sf /usr/src/app/sfc /usr/local/bin/sfc

CMD ["sfc"]