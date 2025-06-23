package types

import (
	fmt "fmt"
	"strings"
	"unicode"

	"cosmossdk.io/math"
)

type PRICE int

const (
	// usd based price: BTC/USDT, BTC/USDC
	U_PRICE PRICE = iota
	// btc based price: ETH/BTC, ATOM/BTC
	BTC_PRICE
)

// digitOrZeroCount returns:
// - If n > 1: the number of digits in n
// - If 0 < n <= 1: the number of zeros after the decimal point before the first non-zero digit
func digitOrZeroCount(n math.LegacyDec) math.Int {
	if n.GTE(math.LegacyOneDec()) {
		// Count digits in the integer part
		s := fmt.Sprintf("%d", n.RoundInt().Int64())
		return math.NewInt(int64(len(s)))
	} else if n.GT(math.LegacyZeroDec()) {
		// Convert to string with enough precision
		s := fmt.Sprintf("%.g", n.MustFloat64())
		if idx := strings.Index(s, "."); idx != -1 {
			frac := s[idx+1:]
			count := math.NewInt(0)
			for _, r := range frac {
				if r == '0' {
					count = count.Sub(math.NewInt(1))
				} else if unicode.IsDigit(r) {
					break
				}
			}
			return count
		}
	}
	// Not positive, return 0
	return math.NewInt(0)
}

func PriceTable(n math.Int) []math.LegacyDec {
	base_interval := math.LegacyNewDec(5)
	i := math.LegacyNewDec(1000)
	prices := []math.LegacyDec{}
	adjust := math.LegacyNewDec(1)
	bound := math.LegacyNewDec(100)
	for range n.Int64() {
		adjust = adjust.MulInt64(10)
	}
	for range -n.Int64() {
		adjust = adjust.QuoInt64(10)
	}
	for {
		i = i.Sub(base_interval)
		prices = append(prices, i.Mul(adjust).QuoInt64(1000))
		if i.LTE(bound) {
			break
		}
	}

	return prices
}

func ComputeLiquidatePrice(price math.LegacyDec, typ PRICE) math.LegacyDec {
	adjust := math.LegacyMustNewDecFromStr("0.005") // price precision
	n := digitOrZeroCount(price)
	for range n.Int64() {
		adjust = adjust.MulInt64(10)
	}
	for range -n.Int64() {
		adjust = adjust.QuoInt64(10)
	}

	if typ == BTC_PRICE {
		return price.Quo(adjust).TruncateDec().Mul(adjust)
	}
	return price.Add(adjust).Quo(adjust).TruncateDec().Mul(adjust)
}

func ComputePriceTable(n math.LegacyDec, typ PRICE) []math.LegacyDec {

	filtered := []math.LegacyDec{}
	if typ == BTC_PRICE {
		threshold := n.Mul(math.LegacyMustNewDecFromStr("1.1"))
		c := digitOrZeroCount(threshold)
		prices := PriceTable(c)
		prices = append(PriceTable(c.Add(math.NewInt(1))), prices...)

		for _, price := range prices {
			if price.GTE(threshold) {
				filtered = append(filtered, price)
			}
		}
	} else if typ == U_PRICE {
		threshold := n.Mul(math.LegacyMustNewDecFromStr("0.9"))
		c := digitOrZeroCount(threshold)
		prices := PriceTable(c)
		prices = append(prices, PriceTable(c.Sub(math.NewInt(1)))...)

		for _, price := range prices {
			if price.LTE(threshold) {
				filtered = append(filtered, price)
			}
		}
	}

	return filtered
}
