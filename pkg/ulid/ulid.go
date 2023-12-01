package ulid

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

const (
	LEN26 = 26
	LEN16 = 16
)

type UlidData struct {
	Timestamp  time.Time
	Value      string
	HexValue   string
	UUIDString string
	HexBytes   []byte
}

var (
	uld      UlidData
	err      error
	hexValue []byte
	uldd     ulid.ULID
	uid      uuid.UUID
	mtx      sync.RWMutex
)

func dataFromUlid(ulidID ulid.ULID) *UlidData {
	mtx.Lock()
	defer mtx.Unlock()
	uld.Timestamp = time.UnixMilli(int64(ulidID.Time()))
	uld.Value = ulidID.String()

	uld.HexValue = hex.EncodeToString(ulidID.Bytes())
	uld.HexBytes, err = hex.DecodeString(uld.HexValue)
	if err != nil {
		fmt.Println(err.Error())
	}
	uid, err = uuid.FromBytes(uld.HexBytes)
	if err != nil {
		fmt.Println(err.Error())
	}

	uld.UUIDString = uid.String()

	return &uld
}

// New generates a new ulid.
func NewUlid() *UlidData {
	id := ulid.Make()
	d := dataFromUlid(id)

	return d
}

// Parse tries to parses a base32 or hex uuid into ulid data.
func Parse(str string) (parsed *UlidData, err error) {
	if len(str) != LEN26 {
		str = strings.ReplaceAll(str, "-", "")
		hexValue, err = hex.DecodeString(str)
		if err != nil {
			return nil, err
		}

		if len(hexValue) != LEN16 {
			return nil, errors.New("invalid uuid")
		}
		copy(uldd[:], hexValue)
	} else {
		uldd, err = ulid.Parse(str)
		if err != nil {
			return nil, err
		}
	}

	d := dataFromUlid(uldd)

	return d, nil
}
