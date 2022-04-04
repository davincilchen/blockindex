FROM golang:1.16.5-stretch
WORKDIR /apps
COPY go.mod ./
COPY go.sum ./
COPY ./main.go /apps/
COPY ./models/ /apps/models/
COPY ./ethprocess/ /apps/ethprocess/
RUN go mod download

COPY *.go ./

RUN go build -o /indexer


CMD [ "/indexer" ]