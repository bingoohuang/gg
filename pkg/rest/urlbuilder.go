package rest

import (
	"log"
	"net/url"
	"path"
)

type URL struct {
	Base        string
	SubPaths    []string
	QueryValues url.Values
}

func NewURL(base string) URL {
	return URL{Base: base, QueryValues: make(url.Values)}
}

func (u URL) QueryMap(m map[string]string) URL {
	for k, v := range m {
		u.QueryValues.Add(k, v)
	}

	return u
}

func (u URL) Query(k, v string, kvs ...string) URL {
	u.QueryValues.Add(k, v)

	for i := 0; i+1 < len(kvs); i += 2 {
		u.QueryValues.Add(kvs[i], kvs[i+1])
	}

	return u
}

func (u URL) Paths(paths ...string) URL {
	u.SubPaths = append(u.SubPaths, paths...)
	return u
}

func (u URL) String() string {
	s, err := u.Build()
	if err != nil {
		log.Fatalf("build URL from base %s failed: %v", u.Base, err)
	}

	return s
}

func (u URL) Build() (string, error) {
	base := FixURI(u.Base)
	if !base.OK() {
		return "", base.Err
	}

	b := base.Data

	p2 := append([]string{b.Path}, u.SubPaths...)
	b.Path = path.Join(p2...)

	q := b.Query()
	for k, v := range u.QueryValues {
		for _, vi := range v {
			q.Add(k, vi)
		}
	}
	b.RawQuery = q.Encode()
	return b.String(), nil
}
