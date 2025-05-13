package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

const crlf = "\r\n"

var specialChars = map[rune]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'|':  true,
	'~':  true,
}

func (h Headers) Get(key string) (string, bool) {
	loweredKey := strings.ToLower(key)
	v, ok := h[loweredKey]
	return v, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	header := strings.TrimSpace(string(data[:idx]))

	colonIdx := 0
	for i, char := range header {
		if char == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == 0 || header[colonIdx-1] == ' ' {
		return 0, false, fmt.Errorf("malformed field-name in header")
	}

	fieldName := strings.ToLower(header[:colonIdx])
	for _, char := range fieldName {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !specialChars[char] {
			return 0, false, fmt.Errorf("invalid char in field name")
		}
	}

	fieldValue := strings.TrimSpace(header[colonIdx+1:])

	v, ok := h[fieldName]
	if ok {
		fieldValue = v + ", " + fieldValue
	}

	h[fieldName] = fieldValue

	return idx + 2, false, nil
}
