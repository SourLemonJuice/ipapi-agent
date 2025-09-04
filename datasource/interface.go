package datasource

import "github.com/SourLemonJuice/ipapi-agent/respstruct"

type Interface interface {
	DoRequest(addr string) error
	Fill(resp *respstruct.Query) error
}
