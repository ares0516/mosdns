package netrange

import (
	"errors"
	"fmt"
	"net/netip"
)

var (
	ErrNotSorted    = errors.New("range list is not sorted")
	ErrNotSupported = errors.New("addr is not supported")
	ErrInvalidAddr  = errors.New("addr is invalid")
	ErrInvalidOpt   = errors.New("Invalid Option")
	ErrInvalidRange = errors.New("Invalid Net Range")
	RangeWholeIPv4  = string("0.0.0.0-255.255.255.255")
	RangeWholeIPv6  = string("::-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
)

type IpRange struct {
	first netip.Addr
	last  netip.Addr
}

type RangeList struct {
	enable bool
	e      []IpRange
	sorted bool
}

func NewRangeList() *RangeList {
	return &RangeList{
		e: make([]IpRange, 0),
	}
}

// func (rlist *RangeList) Disable() {
// 	rlist.enable = false
// }

// func (rlist *RangeList) Enable() {
// 	rlist.enable = true
// }

func (rlist *RangeList) Append(newIpRange ...IpRange) {
	tmp := make([]IpRange, 0)
	for _, n := range newIpRange {
		if n.first.Is4() && n.last.Is4() {
			tmp = append(tmp, n)
		}
	}
	rlist.e = append(rlist.e, tmp...)
	rlist.sorted = false
}

func (rlist *RangeList) Sort() {
	if rlist.sorted {
		return
	}

	// merge

	rlist.sorted = true
}

func (rlist *RangeList) Len() int {
	return len(rlist.e)
}

func (rlist *RangeList) Match(addr netip.Addr) (bool, error) {
	return rlist.Contains(addr)
}

func (rlist *RangeList) Contains(addr netip.Addr) (bool, error) {
	if !rlist.sorted {
		return false, ErrNotSorted
	}
	if !addr.IsValid() {
		return false, ErrNotSupported
	}

	if addr.Is6() && !addr.Is4In6() {
		return false, ErrNotSupported
		// TODO: support IPv6
	}

	client := addr.Unmap()
	fmt.Printf("--------------------addr[%s][%d]\n", client.String(), rlist.Len())
	i, j := 0, len(rlist.e)
	for i < j {
		first := rlist.e[i].first
		last := rlist.e[i].last
		fmt.Printf("[%s]~[%s]\n", first.String(), last.String())
		if (client.Compare(first) >= 0) && (client.Compare(last) <= 0) {
			fmt.Printf("--------matched !\n")
			return true, nil
		}
		i = i + 1
	}

	return false, nil
}
