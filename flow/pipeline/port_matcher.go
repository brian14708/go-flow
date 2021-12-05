package pipeline

import (
	"strings"
)

type portMatcher struct {
	ports  []string
	copied bool
}

func newPortMatcher(ports []string) *portMatcher {
	return &portMatcher{ports, false}
}

// match pattern and append to `matches`, also remove from portMatcher
func (p *portMatcher) appendMatch(matches []string, pattern string) ([]string, bool) {
	if !p.copied {
		tmp := make([]string, len(p.ports))
		copy(tmp, p.ports)
		p.ports = tmp
		p.copied = true
	}

	suffixOnly := false
	if strings.IndexByte(pattern, ':') == -1 {
		suffixOnly = true
		pattern = ":" + pattern
	}

	matchAny := false
	n := 0
	for _, port := range p.ports {
		match := false
		if suffixOnly {
			match = strings.HasSuffix(port, pattern)
		} else {
			match = (port == pattern)
		}
		if match {
			matches = append(matches, port)
			matchAny = true
		} else {
			p.ports[n] = port
			n++
		}
	}
	p.ports = p.ports[:n]
	return matches, matchAny
}

func (p *portMatcher) remaining() []string {
	return p.ports
}
