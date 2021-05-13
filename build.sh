
go build -o bin/dune cmd/dune/*.go
GOOS=darwin GOARCH=amd64 go build -o bin/dune-mac cmd/dune/*.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/dune.exe cmd/dune/*.go

# sudo apt-get install mingw-w64
# GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/dune.exe dune/main.go
# GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i586-mingw32msvc-gcc go build -o bin/dune.es dune/main.go
