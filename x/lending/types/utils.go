package types

import (
	fmt "fmt"
	"strings"
	"unicode"

	sdkmath "cosmossdk.io/math"
)

// digitOrZeroCount returns:
// - If n > 1: the number of digits in n
// - If 0 < n <= 1: the number of zeros after the decimal point before the first non-zero digit
func digitOrZeroCount(n sdkmath.LegacyDec) sdkmath.Int {
	if n.GTE(sdkmath.LegacyOneDec()) {
		// Count digits in the integer part
		s := fmt.Sprintf("%d", n.RoundInt().Int64())
		return sdkmath.NewInt(int64(len(s)))
	} else if n.GT(sdkmath.LegacyZeroDec()) {
		// Convert to string with enough precision
		s := fmt.Sprintf("%.g", n.MustFloat64())
		if idx := strings.Index(s, "."); idx != -1 {
			frac := s[idx+1:]
			count := sdkmath.NewInt(0)
			for _, r := range frac {
				if r == '0' {
					count = count.Sub(sdkmath.NewInt(1))
				} else if unicode.IsDigit(r) {
					break
				}
			}

			return count
		}
	}

	// Not positive, return 0
	return sdkmath.NewInt(0)
}
