package toml

import "github.com/BurntSushi/toml"

type TOML map[string]interface{}

func NewTOMLParser() TOML {
	parser := make(TOML)
	return parser
}

func (t TOML) Marshal(m map[string]interface{}) ([]byte, error) {
	return nil, nil
}

func (t TOML) Unmarshal(b []byte) (map[string]interface{}, error) {
	if err := toml.Unmarshal(b, &t); err != nil {
		return nil, err
	}

	return t, nil
}
