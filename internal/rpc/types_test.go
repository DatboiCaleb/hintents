// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"math"
	"math/big"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ----------------------------------------------------------------

func boolPtr(b bool) *bool           { return &b }
func u32Ptr(v uint32) *xdr.Uint32    { u := xdr.Uint32(v); return &u }
func i32Ptr(v int32) *xdr.Int32      { i := xdr.Int32(v); return &i }
func u64Ptr(v uint64) *xdr.Uint64    { u := xdr.Uint64(v); return &u }
func i64Ptr(v int64) *xdr.Int64      { i := xdr.Int64(v); return &i }
func strPtr(s string) *xdr.ScString  { ss := xdr.ScString(s); return &ss }
func symPtr(s string) *xdr.ScSymbol  { ss := xdr.ScSymbol(s); return &ss }
func bytesPtr(b []byte) *xdr.ScBytes { sb := xdr.ScBytes(b); return &sb }

func mustBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid big.Int string: " + s)
	}
	return n
}

// ---- Requirement 1: primitive types -----------------------------------------

func TestScValToGoValue_Void(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvVoid}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestScValToGoValue_BoolTrue(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: boolPtr(true)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, true, got)
}

func TestScValToGoValue_BoolFalse(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: boolPtr(false)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, false, got)
}

func TestScValToGoValue_U32(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(42)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint32(42), got)
}

func TestScValToGoValue_U32_Max(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(math.MaxUint32)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint32(math.MaxUint32), got)
}

func TestScValToGoValue_I32_Negative(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: i32Ptr(-1)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int32(-1), got)
}

func TestScValToGoValue_I32_Min(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: i32Ptr(math.MinInt32)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int32(math.MinInt32), got)
}

func TestScValToGoValue_U64(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU64, U64: u64Ptr(math.MaxUint64)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint64(math.MaxUint64), got)
}

func TestScValToGoValue_I64_Negative(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: i64Ptr(math.MinInt64)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int64(math.MinInt64), got)
}

func TestScValToGoValue_Timepoint(t *testing.T) {
	tp := xdr.TimePoint(12345)
	v := xdr.ScVal{Type: xdr.ScValTypeScvTimepoint, Timepoint: &tp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint64(12345), got)
}

func TestScValToGoValue_Duration(t *testing.T) {
	dur := xdr.Duration(99999)
	v := xdr.ScVal{Type: xdr.ScValTypeScvDuration, Duration: &dur}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint64(99999), got)
}

// ---- Requirement 2: 128/256-bit integers ------------------------------------

func TestScValToGoValue_U128_Zero(t *testing.T) {
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvU128,
		U128: &xdr.UInt128Parts{Hi: 0, Lo: 0},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_U128_Max(t *testing.T) {
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvU128,
		U128: &xdr.UInt128Parts{Hi: math.MaxUint64, Lo: math.MaxUint64},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	// 2^128 - 1
	expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	assert.Equal(t, 0, got.(*big.Int).Cmp(expected))
}

func TestScValToGoValue_I128_NegativeOne(t *testing.T) {
	// -1 in two's complement 128-bit: Hi = -1 (all ones), Lo = MaxUint64
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvI128,
		I128: &xdr.Int128Parts{Hi: -1, Lo: math.MaxUint64},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Cmp(big.NewInt(-1)))
}

func TestScValToGoValue_I128_MinValue(t *testing.T) {
	// -2^127: Hi = math.MinInt64, Lo = 0
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvI128,
		I128: &xdr.Int128Parts{Hi: math.MinInt64, Lo: 0},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	// -2^127
	expected := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
	assert.Equal(t, 0, got.(*big.Int).Cmp(expected))
}

func TestScValToGoValue_I128_Positive(t *testing.T) {
	// 1: Hi = 0, Lo = 1
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvI128,
		I128: &xdr.Int128Parts{Hi: 0, Lo: 1},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Cmp(big.NewInt(1)))
}

func TestScValToGoValue_U256_Zero(t *testing.T) {
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvU256,
		U256: &xdr.UInt256Parts{HiHi: 0, HiLo: 0, LoHi: 0, LoLo: 0},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_U256_Max(t *testing.T) {
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvU256,
		U256: &xdr.UInt256Parts{
			HiHi: math.MaxUint64,
			HiLo: math.MaxUint64,
			LoHi: math.MaxUint64,
			LoLo: math.MaxUint64,
		},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	// 2^256 - 1
	expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	assert.Equal(t, 0, got.(*big.Int).Cmp(expected))
}

func TestScValToGoValue_I256_NegativeOne(t *testing.T) {
	// -1 in 256-bit two's complement
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvI256,
		I256: &xdr.Int256Parts{
			HiHi: -1,
			HiLo: math.MaxUint64,
			LoHi: math.MaxUint64,
			LoLo: math.MaxUint64,
		},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Cmp(big.NewInt(-1)))
}

func TestScValToGoValue_I256_MinValue(t *testing.T) {
	// -2^255: HiHi = MinInt64, rest = 0
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvI256,
		I256: &xdr.Int256Parts{HiHi: math.MinInt64, HiLo: 0, LoHi: 0, LoLo: 0},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	expected := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 255))
	assert.Equal(t, 0, got.(*big.Int).Cmp(expected))
}

// ---- Requirement 3: bytes / string / symbol ---------------------------------

func TestScValToGoValue_Bytes(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	v := xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: bytesPtr(data)}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestScValToGoValue_Bytes_Empty(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: bytesPtr([]byte{})}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, got)
}

func TestScValToGoValue_String(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: strPtr("hello")}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestScValToGoValue_Symbol(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: symPtr("transfer")}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, "transfer", got)
}

// ---- Requirement 4: ScvVec --------------------------------------------------

func TestScValToGoValue_Vec_Empty(t *testing.T) {
	sv := xdr.ScVec{}
	svp := &sv
	v := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &svp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []interface{}{}, got)
}

func TestScValToGoValue_Vec_Primitives(t *testing.T) {
	sv := xdr.ScVec{
		{Type: xdr.ScValTypeScvU32, U32: u32Ptr(1)},
		{Type: xdr.ScValTypeScvU32, U32: u32Ptr(2)},
		{Type: xdr.ScValTypeScvU32, U32: u32Ptr(3)},
	}
	svp := &sv
	v := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &svp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []interface{}{uint32(1), uint32(2), uint32(3)}, got)
}

func TestScValToGoValue_Vec_DeeplyNested(t *testing.T) {
	// Level 3: innermost vec [42]
	inner := xdr.ScVec{{Type: xdr.ScValTypeScvU32, U32: u32Ptr(42)}}
	innerP := &inner
	innerVal := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &innerP}

	// Level 2: mid vec [innerVal]
	mid := xdr.ScVec{innerVal}
	midP := &mid
	midVal := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &midP}

	// Level 1: outer vec [midVal]
	outer := xdr.ScVec{midVal}
	outerP := &outer
	v := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &outerP}

	got, err := ScValToGoValue(v)
	require.NoError(t, err)

	// Unwrap three levels
	l1, ok := got.([]interface{})
	require.True(t, ok)
	require.Len(t, l1, 1)

	l2, ok := l1[0].([]interface{})
	require.True(t, ok)
	require.Len(t, l2, 1)

	l3, ok := l2[0].([]interface{})
	require.True(t, ok)
	require.Len(t, l3, 1)
	assert.Equal(t, uint32(42), l3[0])
}

// ---- Requirement 5: ScvMap --------------------------------------------------

func TestScValToGoValue_Map_Empty(t *testing.T) {
	sm := xdr.ScMap{}
	smp := &sm
	v := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &smp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []ScValMapEntry{}, got)
}

func TestScValToGoValue_Map_Entries(t *testing.T) {
	sm := xdr.ScMap{
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvString, Str: strPtr("key1")},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(100)}},
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvString, Str: strPtr("key2")},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(200)}},
	}
	smp := &sm
	v := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &smp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	entries := got.([]ScValMapEntry)
	require.Len(t, entries, 2)
	assert.Equal(t, "key1", entries[0].Key)
	assert.Equal(t, uint32(100), entries[0].Value)
	assert.Equal(t, "key2", entries[1].Key)
	assert.Equal(t, uint32(200), entries[1].Value)
}

func TestScValToGoValue_Map_OrderPreserved(t *testing.T) {
	sm := xdr.ScMap{
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(3)},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvVoid}},
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(1)},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvVoid}},
		{Key: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: u32Ptr(2)},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvVoid}},
	}
	smp := &sm
	v := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &smp}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	entries := got.([]ScValMapEntry)
	require.Len(t, entries, 3)
	// Order must match insertion order (3, 1, 2)
	assert.Equal(t, uint32(3), entries[0].Key)
	assert.Equal(t, uint32(1), entries[1].Key)
	assert.Equal(t, uint32(2), entries[2].Key)
}

// ---- Requirement 6: error / nonce / contract instance ----------------------

func TestScValToGoValue_Error(t *testing.T) {
	code := xdr.ScErrorCode(1)
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvError,
		Error: &xdr.ScError{
			Type: xdr.ScErrorTypeSceWasmVm,
			Code: &code,
		},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	se := got.(ScValError)
	assert.Equal(t, uint32(xdr.ScErrorTypeSceWasmVm), se.Type)
	assert.Equal(t, uint32(1), se.Code)
}

func TestScValToGoValue_LedgerKeyContractInstance(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestScValToGoValue_LedgerKeyNonce(t *testing.T) {
	nonce := xdr.Int64(42)
	v := xdr.ScVal{
		Type:     xdr.ScValTypeScvLedgerKeyNonce,
		NonceKey: &xdr.ScNonceKey{Nonce: nonce},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int64(42), got)
}

func TestScValToGoValue_ContractInstance_Wasm(t *testing.T) {
	var hash xdr.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	contractHash := xdr.Hash(hash)
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvContractInstance,
		Instance: &xdr.ScContractInstance{
			Executable: xdr.ContractExecutable{
				Type:     xdr.ContractExecutableTypeContractExecutableWasm,
				WasmHash: &contractHash,
			},
		},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	ci := got.(*ScValContractInstance)
	assert.Equal(t, "wasm", ci.ExecutableType)
	assert.Len(t, ci.WasmHash, 64) // 32 bytes hex-encoded = 64 chars
}

func TestScValToGoValue_ContractInstance_StellarAsset(t *testing.T) {
	v := xdr.ScVal{
		Type: xdr.ScValTypeScvContractInstance,
		Instance: &xdr.ScContractInstance{
			Executable: xdr.ContractExecutable{
				Type: xdr.ContractExecutableTypeContractExecutableStellarAsset,
			},
		},
	}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	ci := got.(*ScValContractInstance)
	assert.Equal(t, "stellar-asset", ci.ExecutableType)
	assert.Empty(t, ci.WasmHash)
}

// ---- Requirement 7: nil-pointer safety --------------------------------------

func TestScValToGoValue_NilBool(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, false, got)
}

func TestScValToGoValue_NilU32(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), got)
}

func TestScValToGoValue_NilI32(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int32(0), got)
}

func TestScValToGoValue_NilU64(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU64, U64: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), got)
}

func TestScValToGoValue_NilI64(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}

func TestScValToGoValue_NilU128(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU128, U128: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_NilI128(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_NilU256(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvU256, U256: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_NilI256(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvI256, I256: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, 0, got.(*big.Int).Sign())
}

func TestScValToGoValue_NilBytes(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, got)
}

func TestScValToGoValue_NilString(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestScValToGoValue_NilSymbol(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestScValToGoValue_NilVec(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []interface{}{}, got)
}

func TestScValToGoValue_NilMap(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, []ScValMapEntry{}, got)
}

func TestScValToGoValue_NilNonce(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyNonce, NonceKey: nil}
	got, err := ScValToGoValue(v)
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)
}

// ---- Requirement 8: unsupported type ----------------------------------------

func TestScValToGoValue_UnsupportedType(t *testing.T) {
	v := xdr.ScVal{Type: xdr.ScValType(9999)}
	_, err := ScValToGoValue(v)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported ScVal type")
}
