package imap

type SearchKey string

const (
	SearchKeyFrom    SearchKey = "FROM"
	SearchKeySubject SearchKey = "SUBJECT"
	SearchKeyBody    SearchKey = "BODY"
	SearchKeyText    SearchKey = "TEXT"
)

type SearchItem struct {
	Key SearchKey
	Val string
}

func NewSearchItem(k SearchKey) *SearchItem {
	return &SearchItem{
		Key: k,
	}
}

func (si *SearchItem) SetVal(val string) *SearchItem {
	si.Val = val

	return si
}
