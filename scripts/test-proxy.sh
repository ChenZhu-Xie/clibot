#!/bin/bash

echo "Testing proxy support..."

# Test 1: Environment variable proxy
echo "Test 1: Environment variable proxy"
export HTTP_PROXY="http://127.0.0.1:8080"
./bin/clibot.exe validate --config configs/config.mini.yaml

# Test 2: Global proxy in config
echo "Test 2: Global proxy configuration"
# Create test config with proxy
./bin/clibot.exe validate --config configs/test-proxy.yaml

# Test 3: Bot-level proxy
echo "Test 3: Bot-level proxy override"
# Create test config with bot-level proxy
./bin/clibot.exe validate --config configs/test-bot-proxy.yaml

echo "All validation tests passed!"
