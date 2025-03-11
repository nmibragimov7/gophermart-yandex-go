package moonchecker

import "unicode"

func MoonChecker(number string) bool {
	sum := 0
	alternate := false

	// Проходим строку в обратном порядке
	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]
		if !unicode.IsDigit(rune(ch)) {
			return false // Строка содержит недопустимые символы
		}

		digit := int(ch - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}
