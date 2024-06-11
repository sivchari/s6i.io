//go:build js && wasm

package main

func main() {
	(&Worker{}).Work()
}
