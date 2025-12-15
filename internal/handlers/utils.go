package handlers

// maskLast4 возвращает строку вида ****1234
func maskLast4(s string) string {
	if len(s) <= 4 {
		return "****" + s
	}
	return "****" + s[len(s)-4:]
}
