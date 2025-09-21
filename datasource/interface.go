package datasource

import "github.com/SourLemonJuice/ipapi-agent/resps"

type Interface interface {
	DoRequest(addr string) error
	Fill(resp *resps.Query) error
}
