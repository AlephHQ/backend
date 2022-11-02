package imap

type SpecialCharacter rune

const (
	SpecialCharacterSpace         SpecialCharacter = ' '
	SpecialCharacterStar          SpecialCharacter = '*'
	SpecialCharacterCR            SpecialCharacter = '\r'
	SpecialCharacterLF            SpecialCharacter = '\n'
	SpecialCharacterDoubleQuote   SpecialCharacter = '"'
	SpecialCharacterRespCodeStart SpecialCharacter = '['
	SpecialCharacterRespCodeEnd   SpecialCharacter = ']'
	SpecialCharacterPlus          SpecialCharacter = '+'
	SpecialCharacterListStart     SpecialCharacter = '('
	SpecialCharacterListEnd       SpecialCharacter = ')'
)

var specialChars map[SpecialCharacter]bool

func init() {
	specialChars = make(map[SpecialCharacter]bool)

	specialChars[SpecialCharacterSpace] = true
	specialChars[SpecialCharacterStar] = true
	specialChars[SpecialCharacterCR] = true
	specialChars[SpecialCharacterLF] = true
	specialChars[SpecialCharacterDoubleQuote] = true
	specialChars[SpecialCharacterRespCodeStart] = true
	specialChars[SpecialCharacterRespCodeEnd] = true
	specialChars[SpecialCharacterPlus] = true
	specialChars[SpecialCharacterListStart] = true
	specialChars[SpecialCharacterListEnd] = true
}

func IsSpecialChar(r rune) bool {
	if t, ok := specialChars[SpecialCharacter(r)]; ok && t {
		return true
	}

	return false
}
