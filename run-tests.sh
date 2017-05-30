#!/bin/bash

go test -cover \
  github.com/venicegeo/dg-bf-ia-broker \
  github.com/venicegeo/dg-bf-ia-broker/landsat \
  github.com/venicegeo/dg-bf-ia-broker/planet \
  github.com/venicegeo/dg-bf-ia-broker/tides \
  github.com/venicegeo/dg-bf-ia-broker/util
