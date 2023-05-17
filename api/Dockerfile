FROM golang:1.16.5-stretch
WORKDIR /apps
COPY go.mod ./
COPY go.sum ./
COPY ./main.go /apps/
COPY ./controllers/ /apps/controllers/
COPY ./models/ /apps/models/
COPY ./routes/ /apps/routes/

RUN go mod download

COPY *.go ./

RUN go build -o /api


CMD [ "/api" ]