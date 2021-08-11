package shellwords

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

var (
	ParseEnv      = false
	ParseBacktick = false
)

var envRe = regexp.MustCompile(`\$({[a-zA-Z0-9_]+}|[a-zA-Z0-9_]+)`)

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n':
		return true
	}

	return false
}

func replaceEnv(getenv func(string) string, s string) string {
	if getenv == nil {
		getenv = os.Getenv
	}

	return envRe.ReplaceAllStringFunc(s, func(s string) string {
		s = s[1:]
		if s[0] == '{' {
			s = s[1 : len(s)-1]
		}
		return getenv(s)
	})
}

// Parser keeps state of parsing.
type Parser struct {
	ParseEnv      bool
	ParseBacktick bool
	Position      int
	Dir           string

	// If ParseEnv is true, use this for getenv.
	// If nil, use os.Getenv.
	Getenv func(string) string
}

// NewParser creates a Parser
func NewParser() *Parser {
	return &Parser{
		ParseEnv:      ParseEnv,
		ParseBacktick: ParseBacktick,
	}
}

type state struct {
	escaped, doubleQuoted, singleQuoted, backQuote, dollarQuote, got bool

	args     []string
	buf      string
	backtick string
	pos      int
}

// Parse parses a lines.
func (p *Parser) Parse(line string) ([]string, error) {
	s := state{
		args: make([]string, 0),
		pos:  -1,
	}

loop:
	for i, r := range line {
		if s.escaped {
			s.buf += string(r)
			s.escaped = false
			continue
		}

		if r == '\\' {
			if s.singleQuoted {
				s.buf += string(r)
			} else {
				s.escaped = true
			}
			continue
		}

		if isSpace(r) {
			if s.singleQuoted || s.doubleQuoted || s.backQuote || s.dollarQuote {
				s.buf += string(r)
				s.backtick += string(r)
			} else if s.got {
				if p.ParseEnv {
					va := replaceEnv(p.Getenv, s.buf)
					if va != "" {
						s.args = append(s.args, va)
					}
				} else {
					s.args = append(s.args, s.buf)
				}
				s.buf = ""
				s.got = false
			}
			continue
		}

		switch r {
		case '`':
			if !s.singleQuoted && !s.doubleQuoted && !s.dollarQuote {
				if p.ParseBacktick {
					if s.backQuote {
						out, err := shellRun(s.backtick, p.Dir)
						if err != nil {
							return nil, err
						}
						s.buf = s.buf[:len(s.buf)-len(s.backtick)] + out
					}
					s.backtick = ""
					s.backQuote = !s.backQuote
					continue
				}
				s.backtick = ""
				s.backQuote = !s.backQuote
			}
		case ')':
			if !s.singleQuoted && !s.doubleQuoted && !s.backQuote {
				if p.ParseBacktick {
					if s.dollarQuote {
						out, err := shellRun(s.backtick, p.Dir)
						if err != nil {
							return nil, err
						}
						s.buf = s.buf[:len(s.buf)-len(s.backtick)-2] + out
					}
					s.backtick = ""
					s.dollarQuote = !s.dollarQuote
					continue
				}
				s.backtick = ""
				s.dollarQuote = !s.dollarQuote
			}
		case '(':
			if !s.singleQuoted && !s.doubleQuoted && !s.backQuote {
				if !s.dollarQuote && strings.HasSuffix(s.buf, "$") {
					s.dollarQuote = true
					s.buf += "("
					continue
				} else {
					return nil, errors.New("invalid command line string")
				}
			}
		case '"':
			if !s.singleQuoted && !s.dollarQuote {
				if s.doubleQuoted {
					s.got = true
				}
				s.doubleQuoted = !s.doubleQuoted
				continue
			}
		case '\'':
			if !s.doubleQuoted && !s.dollarQuote {
				if s.singleQuoted {
					s.got = true
				}
				s.singleQuoted = !s.singleQuoted
				continue
			}
		case ';', '&', '|', '<', '>':
			if !(s.escaped || s.singleQuoted || s.doubleQuoted || s.backQuote || s.dollarQuote) {
				if r == '>' && len(s.buf) > 0 {
					if c := s.buf[0]; '0' <= c && c <= '9' {
						i--
						s.got = false
					}
				}
				s.pos = i
				break loop
			}
		}

		s.got = true
		s.buf += string(r)
		if s.backQuote || s.dollarQuote {
			s.backtick += string(r)
		}
	}

	if s.got {
		if p.ParseEnv {
			va := replaceEnv(p.Getenv, s.buf)
			if va != "" {
				s.args = append(s.args, va)
			}
		} else {
			s.args = append(s.args, s.buf)
		}
	}

	if s.escaped || s.singleQuoted || s.doubleQuoted || s.backQuote || s.dollarQuote {
		return nil, errors.New("invalid command line string")
	}

	p.Position = s.pos

	return s.args, nil
}

// Parse parses a lines.
func Parse(line string) ([]string, error) {
	return NewParser().Parse(line)
}
