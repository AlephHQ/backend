package imap

type Message struct {
	UID uint64

	SeqNum uint64

	Flags map[string]bool

	InternalDate string

	Size uint64

	Envelope *Envelope

	Body interface{}

	Text string
}

func NewMessage(uid uint64) *Message {
	return &Message{
		UID:   uid,
		Flags: make(map[string]bool),
	}
}

func (m *Message) SetSeqNum(seqnum uint64) *Message {
	m.SeqNum = seqnum

	return m
}

func (m *Message) SetFlag(flag string) *Message {
	m.Flags[flag] = true

	return m
}

func (m *Message) SetInternalDate(date string) *Message {
	m.InternalDate = date

	return m
}

func (m *Message) SetSize(s uint64) *Message {
	m.Size = s

	return m
}

func (m *Message) SetText(t string) *Message {
	m.Text = t

	return m
}

func (m *Message) SetEnvelope(e *Envelope) *Message {
	m.Envelope = e

	return m
}
