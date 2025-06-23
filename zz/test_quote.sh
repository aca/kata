#!/usr/bin/env bash
# Test what Go's %q produces
cat > test_go_quote.go << 'EOF'
package main
import "fmt"
func main() {
    fmt.Printf("args[0]=%q\n", "hello world")
    fmt.Printf("args[1]=%q\n", "single'quote")
    fmt.Printf("args[2]=%q\n", "double\"quote")
    fmt.Printf("args[3]=%q\n", "newline\nchar")
}
EOF
go run test_go_quote.go
rm -f test_go_quote.go