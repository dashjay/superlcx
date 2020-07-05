set -x;

for line in $(ls middlewares); \
  do awk '{if($0 ~ /^package .*?$/) print "package main"} {if($0 !~ /^package .*?/) print $0}' \
  middlewares/$line/handler.go >> temp.go; \
  go build -o $line.so --buildmode=plugin temp.go; \
  rm temp.go;  \
done