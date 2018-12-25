#!/bin/bash

export _HANDLER=index.handler

run() {
runner /var/runtime/bootstrap
}

echo '{}' | run
echo '{"error": "Got error"}' | run
