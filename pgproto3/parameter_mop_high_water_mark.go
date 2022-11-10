package pgproto3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/jackc/pgx/v5/internal/pgio"
)

type ParameterMopHighWaterMark struct {
	HighWaterMark uint64
}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*ParameterMopHighWaterMark) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *ParameterMopHighWaterMark) Decode(src []byte) error {
	buf := bytes.NewBuffer(src)

	if buf.Len() < 8 {
		return &invalidMessageFormatErr{messageType: "NotificationResponse", details: "too short"}
	}

	dst.HighWaterMark = binary.BigEndian.Uint64(buf.Next(8))
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *ParameterMopHighWaterMark) Encode(dst []byte) []byte {
	dst = append(dst, 'k')
	dst = pgio.AppendUint64(dst, src.HighWaterMark)
	return dst
}

// MarshalJSON implements encoding/json.Marshaler.
func (ps ParameterMopHighWaterMark) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type          string
		HighWaterMark uint64
	}{
		Type:          "ParameterMopHighWaterMark",
		HighWaterMark: ps.HighWaterMark,
	})
}
