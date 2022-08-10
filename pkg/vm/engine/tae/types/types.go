// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/encoding"
	"golang.org/x/exp/constraints"
	"math"
)

type TypeId = types.T

const (
	Type_ANY = types.T_any

	Type_BOOL = types.T_bool

	Type_INT8   = types.T_int8
	Type_INT16  = types.T_int16
	Type_INT32  = types.T_int32
	Type_INT64  = types.T_int64
	Type_UINT8  = types.T_uint8
	Type_UINT16 = types.T_uint16
	Type_UINT32 = types.T_uint32
	Type_UINT64 = types.T_uint64

	Type_FLOAT32 = types.T_float32
	Type_FLOAT64 = types.T_float64

	Type_DATE      = types.T_date
	Type_DATETIME  = types.T_datetime
	Type_TIMESTAMP = types.T_timestamp
	Type_INTERVAL  = types.T_interval

	Type_CHAR    = types.T_char
	Type_VARCHAR = types.T_varchar

	Type_JSON = types.T_json

	Type_DECIMAL64  = types.T_decimal64
	Type_DECIMAL128 = types.T_decimal128

	Type_SEL   = types.T_sel
	Type_TUPLE = types.T_tuple
	Type_BLOB  = types.T_blob
)

type Type = types.Type

type Date = types.Date
type Datetime = types.Datetime
type Timestamp = types.Timestamp
type Decimal64 = types.Decimal64
type Decimal128 = types.Decimal128
type Bytes = types.Bytes
type Null struct{}

// timestamp for transaction : physical time(high 8 bytes) + logical time(low 4 bytes)
type TS [12]byte

func (ts TS) IsEmpty() bool {
	return encoding.DecodeInt64(ts[4:12]) == 0 &&
		encoding.DecodeUint32(ts[:4]) == 0
}

func (ts TS) Equal(rhs TS) bool {
	return ts == rhs
}

func (ts TS) Less(rhs TS) bool {
	return encoding.DecodeInt64(ts[4:12]) < encoding.DecodeInt64(rhs[4:12]) ||
		(encoding.DecodeInt64(ts[4:12]) == encoding.DecodeInt64(rhs[4:12]) &&
			encoding.DecodeUint32(ts[:4]) < encoding.DecodeUint32(rhs[:4]))
}

func (ts TS) LessEq(rhs TS) bool {
	return ts.Less(rhs) || ts.Equal(rhs)
}

func (ts TS) Greater(rhs TS) bool {
	return encoding.DecodeInt64(ts[4:12]) > encoding.DecodeInt64(rhs[4:12]) ||
		(encoding.DecodeInt64(ts[4:12]) == encoding.DecodeInt64(rhs[4:12]) &&
			encoding.DecodeUint32(ts[:4]) > encoding.DecodeUint32(rhs[:4]))

}

func (ts TS) Prev() TS {
	pTs := encoding.DecodeInt64(ts[4:12])
	lTs := encoding.DecodeUint32(ts[:4])

	var rv TS
	s1 := rv[4:12]
	s2 := rv[:4]

	if lTs == 0 {
		copy(s1, encoding.EncodeInt64(pTs-1))
		copy(s2, encoding.EncodeUint32(math.MaxUint32))
		return rv
	}

	copy(s1, encoding.EncodeInt64(pTs))
	copy(s2, encoding.EncodeUint32(lTs-1))
	return rv
}

func (ts TS) Next() TS {
	pTs := encoding.DecodeInt64(ts[4:12])
	lTs := encoding.DecodeUint32(ts[:4])

	var rv TS
	s1 := rv[4:12]
	s2 := rv[:4]

	if lTs == math.MaxUint32 {
		copy(s1, encoding.EncodeInt64(pTs+1))
		copy(s2, encoding.EncodeUint32(0))
		return rv
	}

	copy(s1, encoding.EncodeInt64(pTs))
	copy(s2, encoding.EncodeUint32(lTs+1))
	return rv
}

func (ts TS) GreaterEq(rhs TS) bool {
	return ts.Greater(rhs) || ts.Equal(rhs)
}

func (ts TS) ToString() (s string) {
	s = string(ts[:])
	return
}

func (ts TS) ToSlice() (s []byte) {
	s = ts[:]
	return
}

func StringToTS(s string) (ts TS) {
	copy(ts[:], s)
	return
}

func MaxTs() TS {
	var rv TS
	s1 := rv[4:12]
	s2 := rv[:4]

	copy(s1, encoding.EncodeInt64(math.MaxInt64))
	copy(s2, encoding.EncodeUint32(math.MaxUint32))
	return rv
}

var SystemDBTS TS

func init() {
	copy(SystemDBTS[4:12], encoding.EncodeInt64(1))
	copy(SystemDBTS[:4], encoding.EncodeUint32(0))
}

var CompareDecimal128 = types.CompareDecimal128
var FromClock = types.FromClock
var FromCalendar = types.FromCalendar

type OrderedT interface {
	constraints.Ordered | Date | Datetime | Timestamp
}

type DecimalT interface {
	Decimal64 | Decimal128
}

type FixedSizeT interface {
	bool | constraints.Ordered | Date | Datetime | Timestamp | Decimal64 | Decimal128
}

type VarSizeT interface {
	Bytes
}

func IsNull(v any) bool {
	_, ok := v.(Null)
	return ok
}

func DefaultVal[T any]() T {
	var v T
	return v
}
