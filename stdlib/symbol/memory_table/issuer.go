package memory_table

import (
	"errors"
	"sync/atomic"

	"github.com/arcspace/go-arc-sdk/stdlib/symbol"
)

var ErrIssuerClosed = errors.New("issuer is closed")

// issuer implements symbol.Issuer using badger.DB
type issuer struct {
	nextID   atomic.Uint64
	refCount atomic.Int32
}

func newIssuer(opts TableOpts) (symbol.Issuer, error) {
	iss := &issuer{}
	iss.nextID.Store(symbol.MinIssuedID)
	iss.refCount.Store(1)
	return iss, nil
}

func (iss *issuer) IssueNextID() (symbol.ID, error) {
	nextID := iss.nextID.Add(1)
	if nextID < symbol.MinIssuedID {
		return 0, ErrIssuerClosed
	}
	return symbol.ID(nextID), nil
}

func (iss *issuer) AddRef() {
	iss.refCount.Add(1)
}

func (iss *issuer) Close() error {
	if iss.refCount.Add(-1) > 0 {
		return nil
	}
	return iss.close()
}

func (iss *issuer) close() error {
	iss.nextID.Store(0)
	return nil
}
