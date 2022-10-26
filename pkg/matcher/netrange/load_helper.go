package netrange

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/IrineSistiana/mosdns/v4/pkg/data_provider"
	"github.com/IrineSistiana/mosdns/v4/pkg/utils"
)

type MatcherGroup struct {
	g []Matcher
}

func (m *MatcherGroup) Len() int {
	s := 0
	for _, l := range m.g {
		s += l.Len()
	}
	return s
}

func (m *MatcherGroup) Match(addr netip.Addr) (bool, error) {
	for _, list := range m.g {
		ok, err := list.Match(addr)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type DynamicMatcher struct {
	parseFunc func(in []byte) (*RangeList, error)
	v         atomic.Value
}

func NewDynamicMatcher(parseFunc func(in []byte) (*RangeList, error)) *DynamicMatcher {
	return &DynamicMatcher{parseFunc: parseFunc}
}

func (d *DynamicMatcher) Update(newData []byte) error {
	list, err := d.parseFunc(newData)
	if err != nil {
		return err
	}
	d.v.Store(list)
	return nil
}

func (d *DynamicMatcher) Match(addr netip.Addr) (bool, error) {
	return d.v.Load().(*RangeList).Match(addr)
}

func (d *DynamicMatcher) Len() int {
	return d.v.Load().(*RangeList).Len()
}

func BatchLoadProvider(e []string, dm *data_provider.DataManager) (*MatcherGroup, error) {
	mg := new(MatcherGroup)
	staticMatcher := NewRangeList()
	mg.g = append(mg.g, staticMatcher)
	for _, s := range e {
		if strings.HasPrefix(s, "provider:") {
			providerName := strings.TrimPrefix(s, "provider:")
			providerName, _, _ = strings.Cut(providerName, ":")
			provider := dm.GetDataProvider(providerName)
			if provider == nil {
				return nil, fmt.Errorf("connot find provider %s", providerName)
			}

			var parseFunc func(in []byte) (*RangeList, error)
			parseFunc = func(in []byte) (*RangeList, error) {
				l := NewRangeList()
				if err := LoadFromReader(l, bytes.NewReader(in)); err != nil {
					return nil, err
				}
				l.Sort()
				return l, nil
			}

			m := NewDynamicMatcher(parseFunc)
			if err := provider.LoadAndAddListener(m); err != nil {
				return nil, fmt.Errorf("failed to load data from provider %s, %w", providerName, err)
			}
			mg.g = append(mg.g, m)
		} else {
			if err := LoadFromText(staticMatcher, s); err != nil {
				return nil, fmt.Errorf("failed to load data %s, %w", s, err)
			}
		}
	}
	staticMatcher.Sort()
	return mg, nil
}

func LoadFromReader(l *RangeList, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	// count how many lines we have read.
	lineCounter := 0
	for scanner.Scan() {
		lineCounter++
		s := scanner.Text()
		s = strings.TrimSpace(s)
		s = utils.RemoveComment(s, "#")
		s = utils.RemoveComment(s, " ")
		fmt.Printf("-----line[%d]\n", lineCounter)
		if len(s) == 0 {
			if 1 == lineCounter { // first blank line indicate whole net
				s = RangeWholeIPv4
			} else {
				continue
			}
		}
		fmt.Printf("read[%s]\n", s)
		err := LoadFromText(l, s)
		if err != nil {
			return fmt.Errorf("invalid data at line #%d: %w", lineCounter, err)
		}
	}

	if 0 == lineCounter { // blank file indicate whole net
		LoadFromText(l, RangeWholeIPv4)
	}

	return scanner.Err()

}

func LoadFromText(l *RangeList, s string) error {
	first, last, found := strings.Cut(s, "-")
	if !found {
		last = first
	}
	firstAddr, err := netip.ParseAddr(first)
	if err != nil {
		return err
	}
	lastAddr, err := netip.ParseAddr(last)
	if err != nil {
		return err
	}

	if firstAddr.Compare(lastAddr) > 0 { // 网段合法性校验
		return ErrInvalidRange
	}

	if (firstAddr.Is4() && lastAddr.Is4()) || (firstAddr.Is6() && lastAddr.Is6()) {
		var ipRange IpRange
		ipRange.first = firstAddr
		ipRange.last = lastAddr

		l.Append(ipRange)
		fmt.Printf("--------------------first[%s]  last[%s] [%d]\n", firstAddr.String(), lastAddr.String(), l.Len())
	}
	return nil
}
