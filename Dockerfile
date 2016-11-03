FROM hashicorp/terraform:0.7.1

MAINTAINER Adam Friedman <adam.friedman@itaas.dimensiondata.com>

RUN mkdir -p /config
VOLUME /config

WORKDIR /config

COPY _bin/linux-amd64/terraform-provider-ddcloud /bin
