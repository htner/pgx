package pgtype

import (
	errors "golang.org/x/xerrors"
)

// CompositeFields scans the fields of a composite type into the elements of the CompositeFields value. To scan a
// nullable value use a *CompositeFields. It will be set to nil in case of null.
//
// CompositeFields implements EncodeBinary and EncodeText. However, functionality is limited due to CompositeFields not
// knowing the PostgreSQL schema of the composite type. Prefer using a registered CompositeType.
type CompositeFields []interface{}

func (cf CompositeFields) DecodeBinary(ci *ConnInfo, src []byte) error {
	if len(cf) == 0 {
		return errors.Errorf("cannot decode into empty CompositeFields")
	}

	if src == nil {
		return errors.Errorf("cannot decode unexpected null into CompositeFields")
	}

	scanner := NewCompositeBinaryScanner(ci, src)

	for _, f := range cf {
		scanner.ScanValue(f)
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

func (cf CompositeFields) DecodeText(ci *ConnInfo, src []byte) error {
	if len(cf) == 0 {
		return errors.Errorf("cannot decode into empty CompositeFields")
	}

	if src == nil {
		return errors.Errorf("cannot decode unexpected null into CompositeFields")
	}

	scanner := NewCompositeTextScanner(ci, src)

	for _, f := range cf {
		scanner.ScanValue(f)
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

// EncodeText encodes composite fields into the text format. Prefer registering a CompositeType to using
// CompositeFields to encode directly.
func (cf CompositeFields) EncodeText(ci *ConnInfo, buf []byte) ([]byte, error) {
	buf = append(buf, '(')

	fieldBuf := make([]byte, 0, 32)

	for _, f := range cf {
		if f != nil {
			fieldBuf = fieldBuf[0:0]
			if textEncoder, ok := f.(TextEncoder); ok {
				var err error
				fieldBuf, err = textEncoder.EncodeText(ci, fieldBuf)
				if err != nil {
					return nil, err
				}
				if fieldBuf != nil {
					buf = append(buf, QuoteCompositeFieldIfNeeded(string(fieldBuf))...)
				}
			} else {
				dt, ok := ci.DataTypeForValue(f)
				if !ok {
					return nil, errors.Errorf("Unknown data type for %#v", f)
				}

				err := dt.Value.Set(f)
				if err != nil {
					return nil, err
				}

				if textEncoder, ok := dt.Value.(TextEncoder); ok {
					var err error
					fieldBuf, err = textEncoder.EncodeText(ci, fieldBuf)
					if err != nil {
						return nil, err
					}
					if fieldBuf != nil {
						buf = append(buf, QuoteCompositeFieldIfNeeded(string(fieldBuf))...)
					}
				} else {
					return nil, errors.Errorf("Cannot encode text format for %v", f)
				}
			}
		}
		buf = append(buf, ',')
	}

	buf[len(buf)-1] = ')'
	return buf, nil
}

// EncodeBinary encodes composite fields into the binary format. Unlike CompositeType the schema of the destination is
// unknown. Prefer registering a CompositeType to using CompositeFields to encode directly. Because the binary
// composite format requires the OID of each field to be specified the only types that will work are those known to
// ConnInfo.
//
// In particular:
//
// * Nil cannot be used because there is no way to determine what type it.
// * Integer types must be exact matches. e.g. A Go int32 into a PostgreSQL bigint will fail.
// * No dereferencing will be done. e.g. *Text must be used instead of Text.
func (cf CompositeFields) EncodeBinary(ci *ConnInfo, buf []byte) ([]byte, error) {
	b := NewCompositeBinaryBuilder(ci, buf)

	for _, f := range cf {
		dt, ok := ci.DataTypeForValue(f)
		if !ok {
			return nil, errors.Errorf("Unknown OID for %#v", f)
		}

		if binaryEncoder, ok := f.(BinaryEncoder); ok {
			b.AppendEncoder(dt.OID, binaryEncoder)
		} else {
			err := dt.Value.Set(f)
			if err != nil {
				return nil, err
			}
			if binaryEncoder, ok := dt.Value.(BinaryEncoder); ok {
				b.AppendEncoder(dt.OID, binaryEncoder)
			} else {
				return nil, errors.Errorf("Cannot encode binary format for %v", f)
			}
		}
	}

	return b.Finish()
}
