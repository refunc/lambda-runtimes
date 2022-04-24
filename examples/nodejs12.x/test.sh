#!/bin/bash

export _HANDLER=lambda_function.lambda_handler

run() {
runner /var/runtime/bootstrap
}

echo '{}' | run
echo '{"error": "Got error"}' | run
