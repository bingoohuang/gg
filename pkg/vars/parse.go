package vars

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Valuer interface {
	Value(name, params, expr string) (any, error)
}

type ValuerHandler func(name, params string) (any, error)

func (f ValuerHandler) Value(name, params string) (any, error) { return f(name, params) }

func (s Subs) Eval(valuer Valuer) (any, error) {
	if len(s) == 1 && s.CountVars() == len(s) {
		v := s[0].(*SubVar)
		return valuer.Value(v.Name, v.Params, v.Expr)
	}

	value := ""
	for _, sub := range s {
		switch v := sub.(type) {
		case *SubTxt:
			value += v.Val
		case *SubVar:
			vv, err := valuer.Value(v.Name, v.Params, v.Expr)
			if err != nil {
				return nil, err
			}
			value += ToString(vv)
		}
	}

	return value, nil
}

type SubTxt struct {
	Val string
}

func (s SubTxt) IsVar() bool { return false }

type SubVar struct {
	Name   string
	Params string
	Expr   string
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
		} else if bracket := PairBracket(s[0]); bracket != nil {
			if rb := strings.IndexByte(s[1:], bracket.Right); rb > 0 {
				fn := s[1:rb]
				s = s[rb+1:]

				subLiteral, subVar := parseName(&fn, &left, bracket)
				if subLiteral != nil {
					subs = append(subs, subLiteral)
				}
				if subVar != nil {
					subs = append(subs, subVar)
				}
			}
		} else {
			subLiteral, subVar := parseName(&s, &left, bracket)
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

type Bracket struct {
	Left  byte
	Right byte
}

func PairBracket(left byte) *Bracket {
	switch left {
	case '{':
		return &Bracket{Left: '{', Right: '}'}
	case '[':
		return &Bracket{Left: '[', Right: ']'}
	case '#', '%', '`':
		return &Bracket{Left: left, Right: left}
	case '<':
		return &Bracket{Left: '<', Right: '>'}
	}
	return nil
}

func parseName(s, left *string, bracket *Bracket) (subLiteral, subVar Sub) {
	original := *s
	name := ""
	offset := 0
	for i, r := range *s {
		if !validNameRune(r) {
			name = (*s)[:i]
			break
		}
		offset += utf8.RuneLen(r)
	}

	nonParam := name == "" && offset == len(*s)
	if nonParam {
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
				sv.Expr = wrap(original[:rb+1], bracket)
				return
			}
		}
	}

	*s = (*s)[offset:]
	sv.Expr = wrap(original[:offset], bracket)

	return
}

func wrap(s string, bracket *Bracket) string {
	if bracket != nil {
		return "@" + string(bracket.Left) + s + string(bracket.Right)
	}

	return "@" + s
}

func validNameRune(r int32) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) ||
		r == '_' || r == '-' || r == '.'
}

func ToString(value any) string {
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
