VERSION --use-host-command 0.6
FROM betch/godnd:1.18
HOST trix.meh 127.0.0.1
WORKDIR /build

deps:
  COPY --dir cmd matrix ./
  COPY main.go go.mod go.sum ./
  RUN apt update && apt install -y build-essential libolm-dev sqlite3
  RUN go mod download

build:
  FROM +deps
  RUN go build -o trix
  SAVE ARTIFACT trix AS LOCAL build/trix

test:
  FROM +deps
  COPY +build/trix ./
  COPY --dir trixtest ./
  COPY main_test.go ./
  WITH DOCKER --pull betch/trixtest:latest
    RUN docker run -d -p 8008:8008 betch/trixtest:latest && go test -v
  END

all:
  BUILD +build
  BUILD +test
