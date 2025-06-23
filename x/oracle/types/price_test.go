package types_test

import (
	fmt "fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/sideprotocol/side/x/oracle/types"
)

func TestComputePrice(t *testing.T) {

	testCases := []math.LegacyDec{
		math.LegacyNewDec(99999),
		math.LegacyNewDec(9999),
		math.LegacyNewDec(999),
		math.LegacyNewDec(5),
		math.LegacyMustNewDecFromStr("0.0456"),
		math.LegacyMustNewDecFromStr("0.0007860"),
		math.LegacyMustNewDecFromStr("0.9"),
		math.LegacyMustNewDecFromStr("100"),
		math.LegacyMustNewDecFromStr("10"),
		math.LegacyMustNewDecFromStr("1"),
		math.LegacyNewDec(59999),
		math.LegacyNewDec(199),
	}
	for _, p := range testCases {
		// x
		pt := types.ComputePriceTable(p, false)
		fmt.Printf("n: %g, len, %d pt: %v\n", p.MustFloat64(), len(pt), pt)

	}
	t.Fatal("stop here to test digitOrZeroCount function manually")
}

func TestComputeLiquidatePrice(t *testing.T) {

	testCases := []math.LegacyDec{
		math.LegacyNewDec(99299),
		math.LegacyNewDec(929),
		math.LegacyNewDec(919),
		math.LegacyNewDec(5),
		math.LegacyMustNewDecFromStr("0.0456"),
		math.LegacyMustNewDecFromStr("0.0007860"),
		math.LegacyMustNewDecFromStr("0.9"),
		math.LegacyMustNewDecFromStr("89"),
		math.LegacyMustNewDecFromStr("7.2"),
		math.LegacyMustNewDecFromStr("1"),
		math.LegacyNewDec(59399),
		math.LegacyNewDec(159),
	}
	for _, p := range testCases {
		// x
		pt := types.ComputeLiquidatePrice(p, types.U_PRICE)
		fmt.Printf("n: %g, pt: %v\n", p.MustFloat64(), pt)

	}
	t.Fatal("stop here to test digitOrZeroCount function manually")
}

func TestComputePriceUp(t *testing.T) {

	testCases := []math.LegacyDec{
		math.LegacyNewDec(99999),
		math.LegacyNewDec(9999),
		math.LegacyNewDec(999),
		math.LegacyNewDec(5),
		math.LegacyMustNewDecFromStr("0.0456"),
		math.LegacyMustNewDecFromStr("0.0007860"),
		math.LegacyMustNewDecFromStr("0.9"),
		math.LegacyMustNewDecFromStr("100"),
		math.LegacyMustNewDecFromStr("10"),
		math.LegacyMustNewDecFromStr("1"),
		math.LegacyNewDec(59999),
		math.LegacyNewDec(199),
	}
	for _, p := range testCases {
		// x
		pt := types.ComputePriceTable(p, types.U_PRICE)
		fmt.Printf("n: %g, len, %d pt: %v\n", p.MustFloat64(), len(pt), pt)

	}
	t.Fatal("stop here to test digitOrZeroCount function manually")
}
