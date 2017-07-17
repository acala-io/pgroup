#!/bin/bash

function clean_up() {
  (>&2 echo "stderr")
  exit 0;
}

trap clean_up SIGHUP SIGINT SIGTERM

while true; do
  echo "stdout"
  sleep 0.1
done
