# These are some common variables for Make

IMG_TAG ?= latest

# Image URL to use local building image targets
CONTAINER_IMG ?= ghcr.io/vshn/odootools:$(IMG_TAG)

PREVIEW_IMG ?= registry.cloudscale-lpg-2.appuio.cloud/vshn-odoo-test/odootools:$(IMG_TAG)

# This is a key used to encrypt cookies. Generate a new one with `openssl rand -base64 32`
LOCAL_SECRET_KEY ?= rQKkLcSZ1I5Skruo+jDRLK4YjFsIKbX1YmPFMAxKbWI=

ENV ?= test

DOCKER ?= $(shell which podman &>/dev/null && echo podman || echo docker)
