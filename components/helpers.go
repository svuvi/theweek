package components

import (
	"bytes"
	"log"
	"strings"

	"github.com/yuin/goldmark"
)

// String is the result string, bool indicates if it was trimmed
func trimToRowsOrChars(s string, maxRows, maxChars int) (string, bool) {
	rows := strings.Split(s, "\n")
	var trimmedRows []string
	charCount := 0
	trimmed := false

	for _, row := range rows {
		if charCount+len(row) > maxChars {
			trimmedRows = append(trimmedRows, row[:maxChars-charCount])
			trimmed = true
			break
		}

		trimmedRows = append(trimmedRows, row)
		charCount += len(row) + 1 // +1 for a newline char

		if len(trimmedRows) >= maxRows {
			trimmed = true
			break
		}
	}

	return strings.Join(trimmedRows, "\n"), trimmed
}

func mdStringToHTML(md string) string {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		log.Println("Error when parsing markdown string:\n", err, "\nThe string:\n", md)
	}

	return buf.String()
}
