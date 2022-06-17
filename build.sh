CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
tar -zcvf gsekiro-linux-arm64.tar.gz example gsekiro config.yaml && rm gsekiro
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build .
tar -zcvf gsekiro-windows-arm64.tar.gz example gsekiro.exe config.yaml && rm gsekiro.exe
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build .
tar -zcvf gsekiro-mac-arm64.tar.gz example gsekiro config.yaml && rm gsekiro