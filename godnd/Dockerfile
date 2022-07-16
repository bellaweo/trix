FROM golang:1.18-bullseye
WORKDIR /build

RUN apt update && apt upgrade -y && apt install -y jq
RUN REL=$(curl -fsSL https://api.github.com/repos/earthly/earthly/releases/latest | jq -r '.tag_name') && \
  curl -o dockerd-install.sh -fsSL https://raw.githubusercontent.com/earthly/earthly/${REL}/buildkitd/docker-auto-install.sh && \
  chmod 755 dockerd-install.sh && ./dockerd-install.sh
