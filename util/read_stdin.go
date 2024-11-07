package util

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// Reads all text from standard input and returns it as a string.
//
// Returns an empty string if stdin is empty. Returns an error if there are any
// issues reading from stdin.
func ReadStdin() (string, error) {
	// Check if stdin has data available
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// If there's no data in stdin, return empty string
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}

	// Read from stdin using a buffered reader for efficiency
	reader := bufio.NewReader(os.Stdin)
	var buffer bytes.Buffer

	// Read chunks until EOF
	for {
		chunk, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		buffer.Write(chunk)

		if err == io.EOF {
			break
		}
	}

	return buffer.String(), nil
}
