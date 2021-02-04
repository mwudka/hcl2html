public/hcl_wasm.wasm: go/go.mod go/go.sum go/main.go
	(cd go; GOOS=js GOARCH=wasm go build -o ../public/hcl_wasm.wasm main.go)

