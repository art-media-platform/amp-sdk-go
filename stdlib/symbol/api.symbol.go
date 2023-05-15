package symbol

import (
	"github.com/arcspace/go-arc-sdk/stdlib/generics"
)

// ID is a persistent integer value associated with an immutable string or buffer value.
// For convenience, Table's interface methods accept and return uint64 types, but these are understood to be of type ID.
// ID == 0 denotes nil or unassigned.
type ID uint64

// IDSz is the byte size of a symbol.ID (big endian)
// The tradeoff is between key bytes idle (wasted) in a massive db and exponentially more IDs available.
// With 5 bytes, an app continuously issuing a new symbol ID every millisecond could do so for 35 years.
const IDSz = 5

// MinIssuedID specifies a minimum ID value for newly issued IDs.]
//
// ID values less than this value are reserved for clients to represent hard-wired or "out of band" meaning.
// "Hard-wired" meaning that Table.SetSymbolID() can be called with IDs less than MinIssuedID without risk
// of an auto-issued ID contending with it.
const MinIssuedID = 1000

type Issuer interface {
	generics.RefCloser

	// Issues the next sequential unique ID, starting at MinIssuedID.
	IssueNextID() (ID, error)
}

// Table stores value-ID pairs, designed for high-performance lookup of an ID or byte string.
// This implementation is intended to handle extreme loads, leveraging:
//   - ID-value pairs are cached once read, offering subsequent O(1) access
//   - Internal value allocations are pooled. The default TableOpts.PoolSz of 16k means
//     thousands of buffers can be issued or read under only a single allocation.
//
// All methods are thread-safe.
type Table interface {
	generics.RefCloser

	// Returns the Issuer being used by this Table (passed via TableOpts.Issuer or auto-created if no TableOpts.Issuer was given)
	// Note that retained references should make use of generics.RefCloser to ensure proper closure.
	Issuer() Issuer

	// Returns the symbol ID previously associated with the given string/buffer value.
	// The given value buffer is never retained.
	//
	// If not found and autoIssue == true, a new entry is created and the new ID returned.
	// Newly issued IDs are always > 0 and use the lower bytes of the returned ID (see type ID comments).
	//
	// If not found and autoIssue == false, 0 is returned.
	GetSymbolID(value []byte, autoIssue bool) ID

	// Associates the given buffer value to the given symbol ID, allowing multiple values to be mapped to a single ID.
	// If ID == 0, then this is the equivalent to GetSymbolID(value, true).
	SetSymbolID(value []byte, ID ID) ID

	// Looks up and appends the byte string associated with the given symbol ID to the given buf.
	// If ID is invalid or not found, nil is returned.
	GetSymbol(ID ID, io []byte) []byte
}

func ReadID(in []byte) (uint64, []byte) {
	ID := ((uint64(in[1]) << 32) |
		(uint64(in[2]) << 24) |
		(uint64(in[3]) << 16) |
		(uint64(in[4]) << 8) |
		(uint64(in[5])))

	return ID, in[6:]
}

func (id *ID) ReadFrom(in []byte) int {
	*id = ID(
		(uint64(in[0]) << 32) |
			(uint64(in[1]) << 24) |
			(uint64(in[2]) << 16) |
			(uint64(in[3]) << 8) |
			(uint64(in[4])))

	return IDSz
}

func AppendID(ID uint64, io []byte) []byte {
	return append(io, // big endian marshal
		byte(ID>>32),
		byte(ID>>24),
		byte(ID>>16),
		byte(ID>>8),
		byte(ID))
}

func (id ID) WriteTo(io []byte) []byte {
	return append(io, // big endian marshal
		byte(uint64(id)>>32),
		byte(uint64(id)>>24),
		byte(uint64(id)>>16),
		byte(uint64(id)>>8),
		byte(id))
}
