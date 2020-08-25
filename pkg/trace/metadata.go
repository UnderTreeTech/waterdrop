package trace

import "strings"

type Metadata struct {
	md map[string][]string
}

func New() *Metadata {
	return &Metadata{
		md: make(map[string][]string),
	}
}

func (md *Metadata) Set(key, val string) {
	key = strings.ToLower(key)
	md.md[key] = append(md.md[key], val)
}

func (md *Metadata) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range md.md {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}
