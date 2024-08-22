package tag

// ID is a signed 24 byte UTC time index in big endian form, with 6 bytes of signed seconds and 10 bytes of fractional precision.
// This means there are 47 bits dedicated for whole seconds => +/- 4.4 million years
//
// This also means (ID[0] >> 16) yields a standard 64-bit Unix UTC timestamp.
type ID [3]uint64

// Literal Token ID / [3]int64 / tag.ID
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
