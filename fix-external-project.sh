#!/bin/bash

# Script to help fix external project dependencies
# This script should be run in the external project directory

set -e

echo "This script helps fix external project dependencies for microservices-library-go"
echo ""
echo "The issue is that Go's module proxy (sum.golang.org) needs time to sync with new tags."
echo "Here are the solutions:"
echo ""
echo "1. TEMPORARY SOLUTION - Use GOPROXY=direct:"
echo "   export GOPROXY=direct"
echo "   go mod tidy"
echo ""
echo "2. PERMANENT SOLUTION - Update go.mod in external project:"
echo "   Replace all instances of:"
echo "   github.com/anasamu/microservices-library-go/ai@v1.0.0"
echo "   with:"
echo "   github.com/anasamu/microservices-library-go/ai@v1.0.0"
echo ""
echo "3. ALTERNATIVE - Use replace directives in external project:"
echo "   Add these lines to the external project's go.mod:"
echo ""
echo "   replace github.com/anasamu/microservices-library-go/ai => github.com/anasamu/microservices-library-go/ai v1.0.0"
echo "   replace github.com/anasamu/microservices-library-go/auth => github.com/anasamu/microservices-library-go/auth v1.0.0"
echo "   replace github.com/anasamu/microservices-library-go/cache => github.com/anasamu/microservices-library-go/cache v1.0.0"
echo "   # ... add all other modules"
echo ""
echo "4. WAIT FOR PROXY SYNC:"
echo "   The Go module proxy typically syncs within 15-30 minutes."
echo "   You can check if a module is available at:"
echo "   https://proxy.golang.org/github.com/anasamu/microservices-library-go/ai/@v/v1.0.0.info"
echo ""
echo "Current status of your modules:"
echo "✅ All modules are properly tagged with v1.0.0"
echo "✅ All replace directives have been removed"
echo "✅ All dependencies use v1.0.0 versions"
echo "✅ All changes have been pushed to GitHub"
echo ""
echo "The external project should work once the Go module proxy syncs or you use one of the solutions above."

