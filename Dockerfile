FROM hashicorp/terraform:0.7.1

RUN mkdir -p /config
VOLUME /config

WORKDIR /config

COPY _bin/linux-amd64/terraform-provider-ddcloud /bin
