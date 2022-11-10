package pgproto3

import (
	"bytes"
	"encoding/json"
)

type WroteXlog struct {
	value byte
}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*WroteXlog) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 1 byte message length.
func (dst *WroteXlog) Decode(src []byte) (err error) {
	buf := bytes.NewBuffer(src)

	if buf.Len() < 1 {
		return &invalidMessageFormatErr{messageType: "NotificationResponse", details: "too short"}
	}
	dst.value, err = buf.ReadByte()
	if err != nil {
		return err
	}
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 1 byte message length.
func (src *WroteXlog) Encode(dst []byte) []byte {
	dst = append(dst, 'x')
	dst = append(dst, src.value)
	return dst
}

// MarshalJSON implements encoding/json.Marshaler.
func (ps WroteXlog) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      string
		WroteXlog byte
	}{
		Type:      "WroteXlog",
		WroteXlog: ps.value,
	})
}
