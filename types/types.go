package types

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate protoc -I. -I../.. --go_out=. --go_opt=paths=source_relative types.proto



