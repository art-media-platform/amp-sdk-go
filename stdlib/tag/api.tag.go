package tag

type Literal struct {
	ID    ID     // deterministic hash of Token -- (token may or may not be included)
	Token string // utf8 human readable exact / canonical glyph or alias of ID -- 64 byte courtesy limit
}

// tag.Value wraps attribute data elements, exposing its "natural" type name and serialization methods.
type Value interface {
	ValuePb

	// Returns the element type name (a scalar tag.Spec).
	TagSpec() Spec

	// Marshals this Value to a buffer, reallocating if needed.
	MarshalToStore(in []byte) (out []byte, err error)

	// Unmarshals and merges value state from a buffer.
	Unmarshal(src []byte) error

	// Creates a default instance of this same Tag type
	New() Value
}

// Serialization abstraction
type ValuePb interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}
