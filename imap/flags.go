package imap

type Flag string

const (
	FlagDeleted  Flag = `\Deleted`
	FlagSeen     Flag = `\Seen`
	FlagAnswered Flag = `\Answered`
	FlagFlagged  Flag = `\Flagged`
	FlagDraft    Flag = `\Draft`
	FlagCatchAll Flag = `\*`
)
