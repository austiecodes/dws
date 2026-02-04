package codec

import (
	"encoding/json"
	"sync"

	"google.golang.org/grpc/encoding"
)

const JSONCodecName = "json"

var registerOnce sync.Once

type jsonCodec struct{}

func (jsonCodec) Name() string {
	return JSONCodecName
}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Register ensures the JSON codec is registered exactly once.
func Register() {
	registerOnce.Do(func() {
		encoding.RegisterCodec(jsonCodec{})
	})
}
