package datasource

import "github.com/SourLemonJuice/ipapi-agent/resps"

type Interface interface {
	// Prepare data for the next use. This won't fill "status" and "message".
	// If failure, response OK with JSON message.
	DoRequest(addr string) error
	// Fill data into the given struct.
	// If failure, response InternalServerError.
	Fill(resp *resps.Query) error
}
