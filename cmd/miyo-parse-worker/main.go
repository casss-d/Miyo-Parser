package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/CassD/miyo"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		title := scanner.Text()
		if err := encoder.Encode(miyo.Parse(title)); err != nil {
			fmt.Fprintf(os.Stderr, "encode parse result: %v\n", err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read title: %v\n", err)
		os.Exit(1)
	}
}
