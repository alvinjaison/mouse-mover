//go:build !windows && !darwin

package main

import "fmt"

func main() {
	fmt.Println("Mouse Mover is a Windows-only application.")
}
