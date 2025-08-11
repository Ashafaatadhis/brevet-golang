package helpers

var angka = []string{"", "Satu", "Dua", "Tiga", "Empat", "Lima", "Enam", "Tujuh", "Delapan", "Sembilan", "Sepuluh", "Sebelas"}

// NumToString is function to convert 10000 to sepuluh ribu
func NumToString(n int) string {
	switch {
	case n < 12:
		return angka[n]
	case n < 20:
		return NumToString(n-10) + " Belas"
	case n < 100:
		return NumToString(n/10) + " Puluh " + NumToString(n%10)
	case n < 200:
		return "Seratus " + NumToString(n-100)
	case n < 1000:
		return NumToString(n/100) + " Ratus " + NumToString(n%100)
	case n < 2000:
		return "Seribu " + NumToString(n-1000)
	case n < 1000000:
		return NumToString(n/1000) + " Ribu " + NumToString(n%1000)
	case n < 1000000000:
		return NumToString(n/1000000) + " Juta " + NumToString(n%1000000)
	default:
		return "Angka terlalu besar"
	}
}
