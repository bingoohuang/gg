package vars

import (
	"fmt"
	"strings"
	"unicode"
)

type Valuer interface {
	Value(name, params string) interface{}
}

func (s Subs) Eval(valuer Valuer) interface{} {
	if len(s) == 1 && s.CountVars() == len(s) {
		v := s[0].(*SubVar)
		return valuer.Value(v.Name, v.Params)
	}

	value := ""
	for _, sub := range s {
		switch v := sub.(type) {
		case *SubTxt:
			value += v.Val
		case *SubVar:
			vv := valuer.Value(v.Name, v.Params)
			value += ToString(vv)
		}
	}

	return value
}

type SubTxt struct {
	Val string
}

func (s SubTxt) IsVar() bool { return false }

type SubVar struct {
	Name   string
	Params string
}

func (s SubVar) IsVar() bool { return true }

type Sub interface {
	IsVar() bool
}

type Subs []Sub

func (s Subs) CountVars() (count int) {
	for _, sub := range s {
		if sub.IsVar() {
			count++
		}
	}

	return
}

func ParseExpr(src string) Subs {
	s := src
	var subs []Sub
	left := ""
	for {
		a := strings.Index(s, "@")
		if a < 0 || a == len(s)-1 {
			left += s
			break
		}

		left += s[:a]

		a++
		s = s[a:]
		if s[0] == '@' {
			s = s[1:]
			left += "@"
		} else if s[0] == '{' {
			if rb := strings.IndexByte(s, '}'); rb > 0 {
				fn := s[1:rb]
				s = s[rb+1:]

				subLiteral, subVar := parseName(&fn, &left)
				if subLiteral != nil {
					subs = append(subs, subLiteral)
				}
				if subVar != nil {
					subs = append(subs, subVar)
				}
			}
		} else {
			subLiteral, subVar := parseName(&s, &left)
			if subLiteral != nil {
				subs = append(subs, subLiteral)
			}
			if subVar != nil {
				subs = append(subs, subVar)
			}
		}
	}

	if left != "" {
		subs = append(subs, &SubTxt{Val: left})
	}

	if Subs(subs).CountVars() == 0 {
		return []Sub{&SubTxt{Val: src}}
	}

	return subs
}

func parseName(s *string, left *string) (subLiteral, subVar Sub) {
	name := ""
	offset := 0
	for i, r := range *s {
		offset = i
		if !(unicode.IsLetter(r) || unicode.Is(unicode.Han, r) || unicode.IsDigit(r) || r == '_' || r == '-') {
			name = (*s)[:i]
			break
		}
	}

	nonParam := false
	if name == "" && offset == len(*s)-1 {
		nonParam = true
		offset++
		name = *s
	}

	if *left != "" {
		subLiteral = &SubTxt{Val: *left}
		*left = ""
	}

	sv := &SubVar{
		Name: name,
	}
	subVar = sv

	if !nonParam && offset > 0 && offset < len(*s) {
		if (*s)[offset] == '(' {
			if rb := strings.IndexByte(*s, ')'); rb > 0 {
				sv.Params = (*s)[offset+1 : rb]
				*s = (*s)[rb+1:]
				return
			}
		}
	}

	*s = (*s)[offset:]

	return
}

func ToString(value interface{}) string {
	switch vv := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", vv)
	case float32, float64:
		return fmt.Sprintf("%f", vv)
	case bool:
		return fmt.Sprintf("%t", vv)
	case string:
		return vv
	default:
		vvv := fmt.Sprintf("%v", value)
		return vvv
	}
}
