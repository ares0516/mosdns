package netrange

import (
	"net/netip"
)

type Matcher interface {
	Match(addr netip.Addr) (bool, error)
	Len() int
}
