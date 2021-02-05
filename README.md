### Overview

This repo contains a simple example application demonstrating how to consume HCL from JavaScript. It consists over a simple HCL-to-HTML templating
library converter in go, plus a small playground page. It exposes [hcl/v2](github.com/hashicorp/hcl/v2) to JS via the power of cross compilation. 

### But how?

The main app's entry point is [index.js](src/index.js). That function just creates monaco (aka VSCode) instances to allow entering an HCL template and JSON data. It
then periodically calls `parse_hcl` with the entered JSON/HCL.

And where is `parse_hcl` defined? Not in JS! Instead, it's implemented in [go/main.go](go/main.go). That function is compiled to WASM by running:

    GOOS=js GOARCH=wasm go build -o ../public/hcl_wasm.wasm main.go

main.go just contains a simple wrapper around the [hcl/v2](github.com/hashicorp/hcl/v2) library that turns HCL blocks into DOM elements. 