package imap

var specialChars map[SpecialCharacter]bool

func init() {
	specialChars = make(map[SpecialCharacter]bool)

	specialChars[SpecialCharacterSpace] = true
	specialChars[SpecialCharacterStar] = true
	specialChars[SpecialCharacterCR] = true
	specialChars[SpecialCharacterLF] = true
	specialChars[SpecialCharacterDoubleQuote] = true
	specialChars[SpecialCharacterOpenBracket] = true
	specialChars[SpecialCharacterCloseBracket] = true
	specialChars[SpecialCharacterPlus] = true
	specialChars[SpecialCharacterOpenParen] = true
	specialChars[SpecialCharacterCloseParen] = true
	specialChars[SpecialCharacterOpenCurly] = true
	specialChars[SpecialCharacterCloseCurly] = true
}
