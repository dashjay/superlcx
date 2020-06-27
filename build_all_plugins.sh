for line in $(ls plugins); \
  do go build -o $line.so --buildmode=plugin plugins/$line/handler.go;  \
done