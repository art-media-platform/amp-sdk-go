package arc

import "github.com/google/uuid"

type UID [16]byte

// Forms an arc.UID explicitly from two uint64 values.
func FormUID(n0, n1 uint64) UID {
	uid := UID{}
	shift := uint(56)
	for i := 0; i < 8; i++ {
		uid[i+0] = byte(n0 >> shift)
		uid[i+8] = byte(n1 >> shift)
		shift -= 8
	}
	return uid
}

// ParseUID decodes s into a UID or returns an error.  Accepted forms:
//   - xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
//   - xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func ParseUID(s string) (UID, error) {
	uid, err := uuid.Parse(s)
	return UID(uid), err
}

// MustParseUID decodes s into a UID or panics -- see ParseUID().
func MustParseUID(s string) UID {
	uid := uuid.MustParse(s)
	return UID(uid)
}

// String returns the string form of uid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx or "" if uuid is zero.
func (uid UID) String() string {
	return uuid.UUID(uid).String()
}
