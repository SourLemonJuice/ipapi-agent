package datasource

import "github.com/SourLemonJuice/ipapi-agent/respstruct"

type Interface interface {
	DoRequest(addr string) error
	IsSuccess() bool
	GetMessage() string
	Fill(resp *respstruct.Query) error
}
