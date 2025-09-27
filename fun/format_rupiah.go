package fun

import "fmt"

// FormatRupiah formats an integer as Indonesian Rupiah, e.g. 1000000 -> "1.000.000"
func FormatRupiah(amount int) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if (n-i)%3 == 0 && i != 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
