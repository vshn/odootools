# These are some common variables for Make

IMG_TAG ?= latest

# Image URL to use all building/pushing image targets
CONTAINER_IMG ?= ghcr.io/vshn/odootools:$(IMG_TAG)
