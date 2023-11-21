package arc

import (
	"crypto"
	"encoding/binary"

	"github.com/google/uuid"
)

type UID [2]uint64

var (
	NilUID         = UID{}
	DevicePlanet   = FormUID(0, 0x01)
	HostPlanet     = FormUID(0, 0x02)
	AppHomePlanet  = FormUID(0, 0x03)
	UserHomePlanet = FormUID(0, 0x04)
)


func StringToUID(str string) UID {
	hash := crypto.MD5.New()
	hash.Write([]byte(str))
	digest := hash.Sum(nil)
	return [2]uint64{
		binary.BigEndian.Uint64(digest[0:8]),
		binary.BigEndian.Uint64(digest[8:16]),
	}
}


// Forms an arc.UID explicitly from two uint64 values.
func FormUID(n0, n1 uint64) UID {
	return UID{uint64(n0), uint64(n1)}
	/*
		uid := UID{}
		shift := uint(56)
		for i := 0; i < 8; i++ {
			uid[i+0] = byte(n0 >> shift)
			uid[i+8] = byte(n1 >> shift)
			shift -= 8
		}
		return uid*/
}

func BytesToUID(b []byte) (uid UID, err error) {
	if len(b) != 16 {
		return NilUID, ErrCode_InvalidUID.Error("invalid UID length")
	}
	uid[0] = binary.BigEndian.Uint64(b[0:8])
	uid[1] = binary.BigEndian.Uint64(b[8:16])
	return uid, nil
}

// ParseUID decodes s into a UID or returns an error.  Accepted forms:
//   - xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//   - {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
//   - xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func ParseUID(s string) (UID, error) {
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
