package codec

type Codec[T any] interface {
	Encode(msg *T) ([]byte, error)
	Decode(data []byte) (*T, error)
}
