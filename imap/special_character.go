package imap

type SpecialCharacter rune

const (
	SpecialCharacterSpace        SpecialCharacter = ' '
	SpecialCharacterStar         SpecialCharacter = '*'
	SpecialCharacterCR           SpecialCharacter = '\r'
	SpecialCharacterLF           SpecialCharacter = '\n'
	SpecialCharacterDoubleQuote  SpecialCharacter = '"'
	SpecialCharacterOpenBracket  SpecialCharacter = '['
	SpecialCharacterCloseBracket SpecialCharacter = ']'
	SpecialCharacterPlus         SpecialCharacter = '+'
	SpecialCharacterOpenParen    SpecialCharacter = '('
	SpecialCharacterCloseParen   SpecialCharacter = ')'
	SpecialCharacterOpenCurly    SpecialCharacter = '{'
	SpecialCharacterCloseCurly   SpecialCharacter = '}'
)

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

func IsSpecialChar(r rune) bool {
	if t, ok := specialChars[SpecialCharacter(r)]; ok && t {
		return true
	}

	return false
}
