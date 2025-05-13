package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/JLee871/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	RequestHeaders headers.Headers
	RequestBody    RequestBody

	state          requestState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestBody []byte

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state:          requestStateInitialized,
		RequestHeaders: headers.NewHeaders(),
		RequestBody:    make([]byte, 0),
	}

	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d", req.state)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	strSlice := strings.Split(requestLineText, crlf)

	requestParts := strings.Split(strSlice[0], " ")

	if len(requestParts) != 3 {
		return nil, 0, fmt.Errorf("invalid number of request parts")
	}

	conn := strings.Split(requestParts[2], "/")
	if len(conn) != 2 {
		return nil, 0, fmt.Errorf("malformed connection protocol / version")
	}

	connType := conn[0]
	version := conn[1]

	if connType != "HTTP" || version != "1.1" {
		return nil, 0, fmt.Errorf("invalid connection protocol / version")
	}

	requestTarget := requestParts[1]

	method := requestParts[0]
	for _, char := range method {
		if !unicode.IsUpper(char) {
			return nil, 0, fmt.Errorf("invalid method")
		}
	}

	newRequestLine := RequestLine{
		HttpVersion:   version,
		RequestTarget: requestTarget,
		Method:        method,
	}

	return &newRequestLine, idx + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil

	case requestStateParsingHeaders:
		n, done, err := r.RequestHeaders.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil

	case requestStateParsingBody:
		value, ok := r.RequestHeaders.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return 0, nil
		}

		contentLen, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("malformed content-length: %s", err)
		}
		r.RequestBody = append(r.RequestBody, data...)
		r.bodyLengthRead += len(data)

		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("error: request body length is longer than content length header value")
		}
		if r.bodyLengthRead == contentLen {
			r.state = requestStateDone
		}

		return len(data), nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("unknown state")
	}
}
