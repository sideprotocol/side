package types

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// runes protocol name
	RunesProtocolName = "runes"

	// runes magic number
	MagicNumber = txscript.OP_13

	// tag indicating that the following are edicts
	TagBody = 0

	// the number of components of each edict
	EdictLen = 4

	// sats in the runes output by default
	RunesOutValue = 546
)

// ParseRunes parses the potential runes protocol from the given tx;
// If no OP_RETURN found, no error returned
// Only support edicts for now
func ParseRunes(tx *wire.MsgTx) ([]*Edict, error) {
	for _, out := range tx.TxOut {
		tokenizer := txscript.MakeScriptTokenizer(0, out.PkScript)
		if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != txscript.OP_RETURN {
			continue
		}

		if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != MagicNumber {
			continue
		}

		var payload []byte

		for tokenizer.Next() {
			if txscript.IsSmallInt(tokenizer.Opcode()) || tokenizer.Opcode() <= txscript.OP_PUSHDATA4 {
				payload = append(payload, tokenizer.Data()...)
			} else {
				return nil, ErrInvalidRunes
			}
		}

		if tokenizer.Err() != nil {
			return nil, ErrInvalidRunes
		}

		return ParseEdicts(tx, payload)
	}

	return nil, nil
}

// ParseEdicts parses the given payload to a set of edicts
func ParseEdicts(tx *wire.MsgTx, payload []byte) ([]*Edict, error) {
	integers, err := DecodeVec(payload)
	if err != nil {
		return nil, err
	}

	if len(integers) < EdictLen+1 || len(integers[1:])%EdictLen != 0 || !integers[0].Equals(uint128.From64(TagBody)) {
		return nil, ErrInvalidRunes
	}

	integers = integers[1:]

	edicts := make([]*Edict, 0)

	for i := 0; i < len(integers); i = i + 4 {
		output := uint32(integers[i+3].Big().Uint64())
		if output > uint32(len(tx.TxOut)) {
			return nil, ErrInvalidRunes
		}

		// actually we only support one edict for now, so delta is unnecessary
		edict := Edict{
			Id: &RuneId{
				Block: integers[i].Big().Uint64(),
				Tx:    uint32(integers[i+1].Big().Uint64()),
			},
			Amount: integers[i+2].String(),
			Output: output,
		}

		edicts = append(edicts, &edict)
	}

	return edicts, nil
}

// ParseEdict parses the given payload to edict
func ParseEdict(payload []byte) (*Edict, error) {
	integers, err := DecodeVec(payload)
	if err != nil {
		return nil, err
	}

	if len(integers) != EdictLen+1 && !integers[0].Equals(uint128.From64(TagBody)) {
		return nil, ErrInvalidRunes
	}

	return &Edict{
		Id: &RuneId{
			Block: integers[1].Big().Uint64(),
			Tx:    uint32(integers[2].Big().Uint64()),
		},
		Amount: integers[3].String(),
		Output: uint32(integers[4].Big().Uint64()),
	}, nil
}

// BuildEdictScript builds the edict script
func BuildEdictScript(runeId string, amount uint128.Uint128, output uint32) ([]byte, error) {
	var id RuneId
	id.MustUnmarshalFromString(runeId)

	edict := Edict{
		Id:     &id,
		Amount: amount.String(),
		Output: output,
	}

	payload := []byte{TagBody}
	payload = append(payload, edict.MustMarshalLEB128()...)

	scriptBuilder := txscript.NewScriptBuilder()
	scriptBuilder.AddOp(txscript.OP_RETURN).AddOp(MagicNumber).AddData(payload)

	return scriptBuilder.Script()
}

func (id *RuneId) ToString() string {
	return fmt.Sprintf("%d:%d", id.Block, id.Tx)
}

func (id *RuneId) FromString(idStr string) error {
	parts := strings.Split(idStr, ":")
	if len(parts) != 2 {
		return ErrInvalidRuneId
	}

	block, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return err
	}

	tx, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return err
	}

	id.Block = block
	id.Tx = uint32(tx)

	return nil
}

func (id *RuneId) MustUnmarshalFromString(s string) {
	err := id.FromString(s)
	if err != nil {
		panic(err)
	}
}

func (id *RuneId) MarshalToBytes() []byte {
	bz := make([]byte, 8+4)

	binary.LittleEndian.PutUint64(bz, id.Block)
	binary.LittleEndian.PutUint32(bz, id.Tx)

	return bz
}

func (id *RuneId) UnmarshalFromBytes(bz []byte) {
	id.Block = binary.LittleEndian.Uint64(bz[:8])
	id.Tx = binary.LittleEndian.Uint32(bz[8:])
}

// Denom returns the corresponding denom for the runes voucher token
func (id *RuneId) Denom() string {
	return fmt.Sprintf("%s/%s", RunesProtocolName, id.ToString())
}

// FromDenom converts the denom to the rune id
func (id *RuneId) FromDenom(denom string) {
	idStr := strings.TrimPrefix(denom, fmt.Sprintf("%s/", RunesProtocolName))

	id.MustUnmarshalFromString(idStr)
}

// MarshalRuneIdFromString marshals the given id string
func MarshalRuneIdFromString(s string) []byte {
	var id RuneId
	id.MustUnmarshalFromString(s)

	return id.MarshalToBytes()
}

// UnmarshalRuneId unmarshals the given bytes to the rune id
func UnmarshalRuneId(bz []byte) RuneId {
	var id RuneId
	id.UnmarshalFromBytes(bz)

	return id
}

func (e *Edict) MustMarshalLEB128() []byte {
	amount := RuneAmountFromString(e.Amount)

	payload := make([]byte, 0)

	payload = append(payload, EncodeUint64(e.Id.Block)...)
	payload = append(payload, EncodeUint32(e.Id.Tx)...)
	payload = append(payload, EncodeUint128(&amount)...)
	payload = append(payload, EncodeUint32(e.Output)...)

	return payload
}

// RuneBalances defines a set of rune balances
type RuneBalances []*RuneBalance

// GetBalance gets the rune balance by id
func (rbs RuneBalances) GetBalance(id string) (int, uint128.Uint128) {
	for i, balance := range rbs {
		if balance.Id == id {
			return i, RuneAmountFromString(balance.Amount)
		}
	}

	return -1, uint128.Zero
}

// Merge merges the another rune balances
func (rbs RuneBalances) Merge(other RuneBalances) RuneBalances {
	var result RuneBalances
	result = append(result, rbs...)

	for _, ob := range other {
		i, b := result.GetBalance(ob.Id)
		if !b.IsZero() {
			result[i].Amount = b.Add(RuneAmountFromString(ob.Amount)).String()
		} else {
			result = append(result, ob)
		}
	}

	return result
}

// Update updates the balance to the specified amount for the given rune id
// The rune balance will be removed if the given amount is zero
// Assume that the given RuneBalances is compact
func (rbs RuneBalances) Update(id string, amount uint128.Uint128) RuneBalances {
	for i, balance := range rbs {
		if balance.Id == id {
			if !amount.IsZero() {
				rbs[i].Amount = amount.String()
			} else {
				rbs = append(rbs[:i], rbs[i+1:]...)
			}

			break
		}
	}

	return rbs
}

// RuneAmountFromString converts the given string to the rune amount
// Panic if any error occurred
func RuneAmountFromString(s string) uint128.Uint128 {
	amount, err := uint128.FromString(s)
	if err != nil {
		panic(err)
	}

	return amount
}

// MarshalRuneAmount marshals the given amount
func MarshalRuneAmount(amount uint128.Uint128) []byte {
	bz := make([]byte, 16)
	amount.PutBytes(bz)

	return bz
}

// MarshalRuneAmountFromString marshals the given amount string
func MarshalRuneAmountFromString(s string) []byte {
	amount := RuneAmountFromString(s)

	return MarshalRuneAmount(amount)
}

// UnmarshalRuneAmount unmarshals the given bytes to the rune amount
func UnmarshalRuneAmount(bz []byte) uint128.Uint128 {
	return uint128.FromBytes(bz)
}
