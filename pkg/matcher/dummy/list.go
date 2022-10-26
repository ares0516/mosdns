package dummy

import (
	"fmt"
	"net/netip"
)

type TextList struct {
	e      []string
	sorted bool
	flag   bool
}

func NewTextList() *TextList {
	return &TextList{
		e:      make([]string, 0),
		flag:   false,
		sorted: false,
	}
}

func (l *TextList) Enable() {
	l.flag = true
}

func (l *TextList) Disable() {
	l.flag = false
}

func (l *TextList) Append(newText ...string) {
	tmp := make([]string, 0)
	for _, n := range newText {
		tmp = append(tmp, n)
	}
	l.e = append(l.e, tmp...)
	l.sorted = false
}

func (l *TextList) Sort() {
	if l.sorted {
		return
	}

	// merge

	l.sorted = true
}

func (l *TextList) Len() int {
	return len(l.e)
}

func (l *TextList) Match(addr netip.Addr) (bool, error) {
	return l.Contains(addr)
}

func (l *TextList) Contains(addr netip.Addr) (bool, error) {
	if l.flag {
		fmt.Printf("dummy-------------true\n")
		return true, nil
	}
	fmt.Printf("dummy-------------false\n")
	return false, nil
}
