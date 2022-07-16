VERSION 0.6
FROM betch/godnd:18
WORKDIR /build

deps:
  COPY --dir cmd matrix ./
  COPY main.go go.mod go.sum ./
  RUN apt update && apt install -y build-essential libolm-dev sqlite3 && go mod download

build:
  FROM +deps
  RUN go build -o trix
  SAVE ARTIFACT trix AS LOCAL build/trix

#utest:
#  FROM +deps
#  COPY main.go .
#  COPY main_test.go .
#  RUN go test

itest:
  FROM +build
  COPY --dir synapse ./
  COPY main_integration_test.go ./
  WITH DOCKER --compose synapse/docker-compose.yaml
    RUN go test
  END

all:
#  BUILD +utest
  BUILD +itest
