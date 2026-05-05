package codec

import "encoding/json"

type JSONCodec[T any] struct{}

func NewCodecJson[T any]() Codec[T] {
	return &JSONCodec[T]{}
}

func (c *JSONCodec[T]) Encode(msg *T) ([]byte, error) {
	return json.Marshal(msg)
}

func (c *JSONCodec[T]) Decode(data []byte) (*T, error) {
	var msg T
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
