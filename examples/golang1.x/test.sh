#!/bin/bash

export _HANDLER=lambda_handler
export GOPROXY=https://goproxy.cn,direct

go build -o lambda_handler lambda_function.go

run() {
runner /var/runtime/bootstrap
}

echo '{"name":"Arvin"}' | run
echo '{"error": "Got error"}' | run
