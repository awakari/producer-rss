#!/bin/bash

export SLUG=ghcr.io/awakari/producer-rss
export VERSION=latest
docker tag awakari/producer-rss "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
