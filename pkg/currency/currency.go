package currency

import (
	"fmt"
	"math"
)

func FormatIDR(cents int) string {
	rupiah := float64(cents) / 100
	return fmt.Sprintf("Rp %s", formatThousands(rupiah))
}

func FormatUSD(cents int) string {
	dollars := float64(cents) / 100
	return fmt.Sprintf("$%.2f", dollars)
}

func ToFloat(cents int) float64 {
	return float64(cents) / 100
}

func FromFloat(f float64) int {
	return int(math.Round(f * 100))
}

func CalcAmount(elapsedSeconds, hourlyRateCents int) int {
	if elapsedSeconds == 0 || hourlyRateCents == 0 {
		return 0
	}
	return int(math.Round(float64(elapsedSeconds) / 3600 * float64(hourlyRateCents)))
}

func CalcTotal(subtotal int, marginPct, taxPct float64) (margin, tax, total int) {
	margin = int(math.Round(float64(subtotal) * marginPct / 100))
	tax = int(math.Round(float64(subtotal+margin) * taxPct / 100))
	total = subtotal + margin + tax
	return
}

func formatThousands(f float64) string {
	intPart := int(f)
	result := ""
	s := fmt.Sprintf("%d", intPart)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}
