package dummy

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
	// lis.
	return false, nil
}

type DynamicMatcher struct {
	parseFunc func(in []byte) (*TextList, error)
	v         atomic.Value
}

func NewDynamicMatcher(parseFunc func(in []byte) (*TextList, error)) *DynamicMatcher {
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
	return d.v.Load().(*TextList).Match(addr)
}

func (d *DynamicMatcher) Len() int {
	return d.v.Load().(*TextList).Len()
}

// dummy provider only load first param
func BatchLoadProvider(e []string, dm *data_provider.DataManager) (*MatcherGroup, error) {
	mg := new(MatcherGroup)
	staticMatcher := NewTextList()
	mg.g = append(mg.g, staticMatcher)
	loadCounter := 0
	for _, s := range e {
		if strings.HasPrefix(s, "provider:") {
			providerName := strings.TrimPrefix(s, "provider:")
			providerName, _, _ = strings.Cut(providerName, ":")
			provider := dm.GetDataProvider(providerName)
			if provider == nil {
				return nil, fmt.Errorf("connot find provider %s", providerName)
			}

			var parseFunc func(in []byte) (*TextList, error)
			parseFunc = func(in []byte) (*TextList, error) {
				l := NewTextList()
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
			loadCounter++
			if loadCounter > 1 {
				continue
			}
			err := LoadFromText(staticMatcher, s)
			if err != nil {
				return nil, fmt.Errorf("failed to load data %s, %w", s, err)
			}
		}
	}
	staticMatcher.Sort()
	return mg, nil
}

func LoadFromReader(l *TextList, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	s := scanner.Text()
	s = strings.TrimSpace(s)
	s = utils.RemoveComment(s, "#")
	s = utils.RemoveComment(s, " ")
	if len(s) == 0 {
		s = string("off")
	}
	fmt.Printf("read----[%s]\n", s)
	err := LoadFromText(l, s)
	if err != nil {
		return fmt.Errorf("invalid data at line #%d: %w", 1, err)
	}
	return scanner.Err()
}

func LoadFromText(l *TextList, s string) error {
	fmt.Printf("---------123---[%s]\n", s)
	if strings.ToLower(s) == string("on") {
		l.flag = true
	}
	if strings.ToLower(s) == string("off") {
		l.flag = false
	}
	//l.Append(s)
	return nil
}
