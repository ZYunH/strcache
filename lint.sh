set -e
go get -u golang.org/x/lint/golint >/dev/null 2>&1
golint
go vet
go mod tidy