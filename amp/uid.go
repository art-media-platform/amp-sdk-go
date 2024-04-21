package amp

import (
	"crypto"
	"encoding/binary"

	"github.com/google/uuid"
)

type UID [2]uint64

var (
	NilUID         = UID{}
	DevicePlanet   = IntsToUID(0, 1)
	HostPlanet     = IntsToUID(0, 2)
	AppHomePlanet  = IntsToUID(0, 3)
	UserHomePlanet = IntsToUID(0, 4)
)

// Forms a UID from a UID string identifier having the form "scope-desc1.scope-desc2. ...""
func StringToUID(str string) UID {
	if str == "" {
		return NilUID
	}
	hash := crypto.MD5.New()
	hash.Write([]byte(str))
	digest := hash.Sum(nil)
	return [2]uint64{
		binary.BigEndian.Uint64(digest[0:8]),
		binary.BigEndian.Uint64(digest[8:16]),
	}
}

// Forms an amp.UID explicitly from two uint64 values.
func IntsToUID(x0, x1 uint64) UID {
	uid := UID{}
	uid[0] = x0
	uid[1] = x1
	return uid
}

func BytesToUID(b []byte) (uid UID, err error) {
	if len(b) > 16 {
		return NilUID, ErrCode_BadValue.Error("invalid UID length")
	}

	uid[0] = binary.BigEndian.Uint64(b[0:8])
	uid[1] = binary.BigEndian.Uint64(b[8:16])
	return uid, nil
}

func (uid UID) AppendTo(dst []byte) []byte {
	dst = binary.BigEndian.AppendUint64(dst, uid[0])
	dst = binary.BigEndian.AppendUint64(dst, uid[1])
	return dst
}

func (uid UID) IsNil() bool {
	return uid[0] == 0 && uid[1] == 0
}

// ParseUID decodes s into a UID or returns an error.  Accepted forms:
//   - xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
//   - xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func ParseUUID(s string) (UID, error) {
	uidBytes, err := uuid.Parse(s)
	if err != nil {
		return NilUID, err
	}
	return BytesToUID(uidBytes[:])
}

// MustParseUID decodes s into a UID or panics -- see ParseUID().
func MustParseUID(s string) UID {
	uidBytes := uuid.MustParse(s)
	uid, _ := BytesToUID(uidBytes[:])
	return uid
}

// String returns the string form of uid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx or "" if uuid is zero.
func (uid UID) ToUUID() (uuid uuid.UUID) {
	binary.BigEndian.PutUint64(uuid[:8], uid[0])
	binary.BigEndian.PutUint64(uuid[8:], uid[1])
	return uuid
}

// String returns the string form of uid: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx or "" if uuid is zero.
func (uid UID) String() string {
	return uid.ToUUID().String()
}
