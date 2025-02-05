package pgproto3

import (
	"bytes"
	"encoding/json"

	"github.com/jackc/pgx/v5/internal/pgio"
)

type MPPQuery struct {
	String string
}

// Frontend identifies this message as sendable by a PostgreSQL frontend.
func (*MPPQuery) Frontend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *MPPQuery) Decode(src []byte) error {
	i := bytes.IndexByte(src, 0)
	if i != len(src)-1 {
		return &invalidMessageFormatErr{messageType: "MPPQuery"}
	}

	dst.String = string(src[:i])

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *MPPQuery) Encode(dst []byte) []byte {
	dst = append(dst, 'W')
	dst = pgio.AppendInt32(dst, int32(4+len(src.String)+1))

	dst = append(dst, src.String...)
	dst = append(dst, 0)

	return dst
}

// MarshalJSON implements encoding/json.Marshaler.
func (src MPPQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   string
		String string
	}{
		Type:   "MPPQuery",
		String: src.String,
	})
}
