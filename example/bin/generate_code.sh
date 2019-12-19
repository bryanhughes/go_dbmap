#!/usr/bin/env bash

mkdir "output/model"

printf "\nCompiling protobufs..."
for d in output/proto/public/*.proto; do
  if [ -f "$d" ]; then
    echo "    $d"
    protoc -I=output/proto --go_out=output/model "$d"
  fi
done

for d in output/proto/test_schema/*.proto; do
  if [ -f "$d" ]; then
    echo "    $d"
    protoc -I=output/proto --go_out=output/model "$d"
  fi
done
