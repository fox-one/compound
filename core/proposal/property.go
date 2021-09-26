package proposal

import (
	"compound/pkg/mtg"
)

type SetProperty struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (s SetProperty) MarshalBinary() ([]byte, error) {
	return mtg.Encode(s.Key, s.Value)
}

func (s *SetProperty) UnmarshalBinary(data []byte) error {
	var key, value string
	if _, err := mtg.Scan(data, &key, &value); err != nil {
		return err
	}

	s.Key = key
	s.Value = value
	return nil
}
