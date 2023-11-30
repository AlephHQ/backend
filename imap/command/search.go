package command

import "aleph/backend/imap"

type Search struct {
	Tag        string
	SearchKeys []*imap.SearchItem
}

func NewCmdSearch() *Search {
	return &Search{
		Tag: getTag(),
	}
}

func (s *Search) AddSearchItem(k *imap.SearchItem) *Search {
	s.SearchKeys = append(s.SearchKeys, k)

	return s
}

func (s *Search) Command() string {
	result := s.Tag + " " + "SEARCH"
	for _, item := range s.SearchKeys {
		result = result + " " + string(item.Key)
		if item.Val != "" {
			result = result + " " + item.Val
		}
	}

	return result
}
