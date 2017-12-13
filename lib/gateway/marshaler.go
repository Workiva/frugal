package gateway

import (
	"encoding/json"
	"io"
)

// Marshaler defines a conversion between a byte sequence and Frugal payloads / fields.
type Marshaler struct{}

// Decoder decodes a byte sequence
type Decoder interface {
	Decode(v interface{}) error
}

// Encoder encodes Frugal payloads / fields into byte sequence.
type Encoder interface {
	Encode(v interface{}) error
}

// ContentType always returns "application/json".
func (*Marshaler) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (m *Marshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v".
func (m *Marshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (m *Marshaler) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (m *Marshaler) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}
