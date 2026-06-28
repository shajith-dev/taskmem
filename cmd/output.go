package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

// printJSON marshals v as indented JSON and writes it to stdout followed by a newline.
func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "%s\n", b)
	return err
}
