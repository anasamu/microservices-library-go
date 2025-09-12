#!/bin/bash

echo "Fixing remaining gateway references..."

# Fix go.mod files
find . -name "go.mod" -type f -exec sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/gateway|github.com/anasamu/microservices-library-go/\1|g' {} \;

# Fix .go files
find . -name "*.go" -type f -exec sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/gateway|github.com/anasamu/microservices-library-go/\1|g' {} \;

# Fix README files
find . -name "README.md" -type f -exec sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/gateway|github.com/anasamu/microservices-library-go/\1|g' {} \;

echo "Gateway references fixed!"

# Add replace directives for main modules in examples
echo "Adding replace directives for main modules..."

# Find all examples/go.mod files and add replace directives
find . -path "*/examples/go.mod" -type f | while read -r file; do
    dir=$(dirname "$file")
    lib_dir=$(dirname "$dir")
    lib_name=$(basename "$lib_dir")
    
    echo "Processing $file for library $lib_name"
    
    # Add replace directive for main module if not exists
    if ! grep -q "replace github.com/anasamu/microservices-library-go/$lib_name =>" "$file"; then
        echo "" >> "$file"
        echo "replace github.com/anasamu/microservices-library-go/$lib_name => .." >> "$file"
    fi
done

echo "Replace directives added!"

# Run go mod tidy in all modules
echo "Running go mod tidy..."
find . -name "go.mod" -type f -exec dirname {} \; | sort -u | while read -r dir; do
    echo "Tidying $dir"
    (cd "$dir" && go mod tidy)
done

echo "All done!"
