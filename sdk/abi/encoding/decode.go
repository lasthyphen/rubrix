package encoding

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	xdr "github.com/davecgh/go-xdr/xdr2"

	"github.com/dapperlabs/flow-go/sdk/abi/types"
	"github.com/dapperlabs/flow-go/sdk/abi/values"
)

// A Decoder decodes XDR-encoded representations of Cadence values.
type Decoder struct {
	dec *xdr.Decoder
}

// Decode returns a Cadence value decoded from its XDR-encoded representation.
//
// This function returns an error if the bytes do not match the given type
// definition.
func Decode(t types.Type, b []byte) (values.Value, error) {
	r := bytes.NewReader(b)
	dec := NewDecoder(r)

	v, err := dec.Decode(t)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// NewDecoder initializes a Decoder that will decode XDR-encoded bytes from the
// given io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{xdr.NewDecoder(r)}
}

// Decode reads XDR-encoded bytes from the io.Reader and decodes them to a
// Cadence value of the given type.
//
// This function returns an error if the bytes do not match the given type
// definition.
func (e *Decoder) Decode(t types.Type) (values.Value, error) {
	switch x := t.(type) {
	case types.Void:
		return e.DecodeVoid()
	case types.Bool:
		return e.DecodeBool()
	case types.String:
		return e.DecodeString()
	case types.Bytes:
		return e.DecodeBytes()
	case types.Address:
		return e.DecodeAddress()
	case types.Int:
		return e.DecodeInt()
	case types.Int8:
		return e.DecodeInt8()
	case types.Int16:
		return e.DecodeInt16()
	case types.Int32:
		return e.DecodeInt32()
	case types.Int64:
		return e.DecodeInt64()
	case types.Uint8:
		return e.DecodeUint8()
	case types.Uint16:
		return e.DecodeUint16()
	case types.Uint32:
		return e.DecodeUint32()
	case types.Uint64:
		return e.DecodeUint64()
	case types.VariableSizedArray:
		return e.DecodeVariableSizedArray(x)
	case types.ConstantSizedArray:
		return e.DecodeConstantSizedArray(x)
	case types.Dictionary:
		return e.DecodeDictionary(x)
	case types.Composite:
		return e.DecodeComposite(x)
	case types.Event:
		return e.DecodeEvent(x)
	default:
		return nil, fmt.Errorf("unsupported type: %T", t)
	}

	return nil, nil
}

// DecodeVoid reads the XDR-encoded representation of a void value.
//
// Void values are skipped by the decoder because they are empty by
// definition, but this function still exists in order to support composite
// types that contain void values.
func (e *Decoder) DecodeVoid() (values.Void, error) {
	// void values are not encoded
	return struct{}{}, nil
}

// DecodeBool reads the XDR-encoded representation of a boolean value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.4
//  RFC Section 4.4 - Boolean
//  Represented as an XDR encoded enumeration where 0 is false and 1 is true
func (e *Decoder) DecodeBool() (values.Bool, error) {
	b, _, err := e.dec.DecodeBool()
	if err != nil {
		return false, err
	}

	return values.Bool(b), nil
}

// DecodeString reads the XDR-encoded representation of a string value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.11
//  RFC Section 4.11 - String
//  Unsigned integer length followed by bytes zero-padded to a multiple of four
func (e *Decoder) DecodeString() (values.String, error) {
	str, _, err := e.dec.DecodeString()
	if err != nil {
		return "", err
	}

	return values.String(str), nil
}

// DecodeBytes reads the XDR-encoded representation of a byte array.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.10
//  RFC Section 4.10 - Variable-Length Opaque Data
//  Unsigned integer length followed by fixed opaque data of that length
func (e *Decoder) DecodeBytes() (values.Bytes, error) {
	b, _, err := e.dec.DecodeOpaque()
	if err != nil {
		return nil, err
	}

	if b == nil {
		b = []byte{}
	}

	return b, nil
}

// DecodeAddress reads the XDR-encoded representation of an address.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.9
//  RFC Section 4.9 - Fixed-Length Opaque Data
//  Fixed-length uninterpreted data zero-padded to a multiple of four
func (e *Decoder) DecodeAddress() (values.Address, error) {
	b, _, err := e.dec.DecodeFixedOpaque(20)
	if err != nil {
		return values.Address{}, err
	}

	return values.BytesToAddress(b), nil
}

// DecodeInt reads the XDR-encoded representation of an arbitrary-precision
// integer value.
//
// An arbitrary-precision integer is encoded as follows:
//   Sign as a byte flag (positive is 1, negative is 0)
//   Absolute value as variable-length big-endian byte array
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.10
//  RFC Section 4.10 - Variable-Length Opaque Data
//  Unsigned integer length followed by fixed opaque data of that length
func (e *Decoder) DecodeInt() (values.Int, error) {
	b, _, err := e.dec.DecodeOpaque()
	if err != nil {
		return values.Int{}, err
	}

	isPositive := b[0] == 1

	i := big.NewInt(0).SetBytes(b[1:])

	if !isPositive {
		i = i.Neg(i)
	}

	return values.NewIntFromBig(i), nil
}

// DecodeInt8 reads the XDR-encoded representation of an int-8 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.1
//  RFC Section 4.1 - Integer
//  32-bit big-endian signed integer in range [-2147483648, 2147483647]
func (e *Decoder) DecodeInt8() (values.Int8, error) {
	i, _, err := e.dec.DecodeInt()
	if err != nil {
		return 0, err
	}

	return values.Int8(i), nil
}

// DecodeInt16 reads the XDR-encoded representation of an int-16 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.1
//  RFC Section 4.1 - Integer
//  32-bit big-endian signed integer in range [-2147483648, 2147483647]
func (e *Decoder) DecodeInt16() (values.Int16, error) {
	i, _, err := e.dec.DecodeInt()
	if err != nil {
		return 0, err
	}

	return values.Int16(i), nil
}

// DecodeInt32 reads the XDR-encoded representation of an int-32 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.1
//  RFC Section 4.1 - Integer
//  32-bit big-endian signed integer in range [-2147483648, 2147483647]
func (e *Decoder) DecodeInt32() (values.Int32, error) {
	i, _, err := e.dec.DecodeInt()
	if err != nil {
		return 0, err
	}

	return values.Int32(i), nil
}

// DecodeInt64 reads the XDR-encoded representation of an int-64 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.5
//  RFC Section 4.5 - Hyper Integer
//  64-bit big-endian signed integer in range [-9223372036854775808, 9223372036854775807]
func (e *Decoder) DecodeInt64() (values.Int64, error) {
	i, _, err := e.dec.DecodeHyper()
	if err != nil {
		return 0, err
	}

	return values.Int64(i), nil
}

// DecodeUint8 reads the XDR-encoded representation of a uint-8 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.2
//  RFC Section 4.2 - Unsigned Integer
//  32-bit big-endian unsigned integer in range [0, 4294967295]
func (e *Decoder) DecodeUint8() (values.Uint8, error) {
	i, _, err := e.dec.DecodeUint()
	if err != nil {
		return 0, err
	}

	return values.Uint8(i), nil
}

// DecodeUint16 reads the XDR-encoded representation of a uint-16 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.2
//  RFC Section 4.2 - Unsigned Integer
//  32-bit big-endian unsigned integer in range [0, 4294967295]
func (e *Decoder) DecodeUint16() (values.Uint16, error) {
	i, _, err := e.dec.DecodeUint()
	if err != nil {
		return 0, err
	}

	return values.Uint16(i), nil
}

// DecodeUint32 reads the XDR-encoded representation of a uint-32 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.2
//  RFC Section 4.2 - Unsigned Integer
//  32-bit big-endian unsigned integer in range [0, 4294967295]
func (e *Decoder) DecodeUint32() (values.Uint32, error) {
	i, _, err := e.dec.DecodeUint()
	if err != nil {
		return 0, err
	}

	return values.Uint32(i), nil
}

// DecodeUint64 reads the XDR-encoded representation of a uint-64 value.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.5
//  RFC Section 4.5 - Unsigned Hyper Integer
//  64-bit big-endian unsigned integer in range [0, 18446744073709551615]
func (e *Decoder) DecodeUint64() (values.Uint64, error) {
	i, _, err := e.dec.DecodeUhyper()
	if err != nil {
		return 0, err
	}

	return values.Uint64(i), nil
}

// DecodeVariableSizedArray reads the XDR-encoded representation of a
// variable-sized array.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.13
//  RFC Section 4.13 - Variable-Length Array
//  Unsigned integer length followed by individually XDR-encoded array elements
func (e *Decoder) DecodeVariableSizedArray(t types.VariableSizedArray) (values.VariableSizedArray, error) {
	size, _, err := e.dec.DecodeUint()
	if err != nil {
		return nil, err
	}

	constantType := types.ConstantSizedArray{
		Size:        int(size),
		ElementType: t.ElementType,
	}

	array, err := e.DecodeConstantSizedArray(constantType)
	if err != nil {
		return nil, err
	}

	return values.VariableSizedArray(array), nil
}

// DecodeConstantSizedArray reads the XDR-encoded representation of a
// constant-sized array.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.12
//  RFC Section 4.12 - Fixed-Length Array
//  Individually XDR-encoded array elements
func (e *Decoder) DecodeConstantSizedArray(t types.ConstantSizedArray) (values.ConstantSizedArray, error) {
	return e.decodeArray(t.ElementType, t.Size)
}

// decodeArray reads the XDR-encoded representation of a constant-sized array.
//
// Reference: https://tools.ietf.org/html/rfc4506#section-4.12
//  RFC Section 4.12 - Fixed-Length Array
//  Individually XDR-encoded array elements
func (e *Decoder) decodeArray(t types.Type, size int) ([]values.Value, error) {
	array := make([]values.Value, size)

	for i := 0; i < size; i++ {
		value, err := e.Decode(t)
		if err != nil {
			return nil, err
		}

		array[i] = value
	}

	return array, nil
}

// DecodeDictionary reads the XDR-encoded representation of a dictionary.
//
// The size of the dictionary is encoded as an unsigned integer, followed by
// the dictionary keys, then elements, each represented as individually
// XDR-encoded array elements.
func (e *Decoder) DecodeDictionary(t types.Dictionary) (values.Dictionary, error) {
	size, _, err := e.dec.DecodeUint()
	if err != nil {
		return nil, err
	}

	keys, err := e.decodeArray(t.KeyType, int(size))
	if err != nil {
		return nil, err
	}

	elements, err := e.decodeArray(t.ElementType, int(size))
	if err != nil {
		return nil, err
	}

	d := make(values.Dictionary, size)

	for i := 0; i < int(size); i++ {
		key := keys[i]
		element := elements[i]

		d[i] = values.KeyValuePair{
			Key:   key,
			Value: element,
		}
	}

	return d, nil
}

// DecodeComposite reads the XDR-encoded representation of a composite value.
//
// A composite is encoded as a fixed-length array of its field values.
func (e *Decoder) DecodeComposite(t types.Composite) (values.Composite, error) {
	fields := make([]values.Value, len(t.FieldTypes))

	for i, fieldType := range t.FieldTypes {
		value, err := e.Decode(fieldType)
		if err != nil {
			return values.Composite{}, err
		}

		fields[i] = value
	}

	return values.Composite{Fields: fields}, nil
}

// DecodeEvent reads the XDR-encoded representation of an event.
//
// An event is encoded as a fixed-length array of its field values.
func (e *Decoder) DecodeEvent(t types.Event) (values.Event, error) {
	fields := make([]values.Value, len(t.FieldTypes))

	for i, field := range t.FieldTypes {
		value, err := e.Decode(field.Type)
		if err != nil {
			return values.Event{}, err
		}

		fields[i] = value
	}

	return values.Event{
		Identifier: t.Identifier,
		Fields:     fields,
	}, nil
}