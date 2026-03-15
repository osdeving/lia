package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("lia-lsp: starting")
	log.Println("LIA Language Server Protocol (LSP) stub. Run 'lia check' for validation.")
	
	// An actual LSP loop would read JSON-RPC from Stdin here.
	// For now, this serves merely as a placeholder required by T7.1.
	os.Exit(0)
}
