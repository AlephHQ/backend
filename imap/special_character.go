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

func IsSpecialChar(r rune) bool {
	if t, ok := specialChars[SpecialCharacter(r)]; ok && t {
		return true
	}

	return false
}
