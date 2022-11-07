package imap

type Address struct {
	Name         string
	AtDomainList interface{}
	Mailbox      string
	Host         string
}

func NewAddress(name, mailbox, host string) *Address {
	return &Address{
		Name:    name,
		Mailbox: mailbox,
		Host:    host,
	}
}

type Envelope struct {
	Date      string
	Subject   string
	From      []*Address
	Sender    []*Address
	ReplyTo   []*Address
	To        []*Address
	CC        []*Address
	BCC       []*Address
	InReplyTo []*Address
	MessageID string
}

func NewEnvelope() *Envelope {
	return &Envelope{
		From:      make([]*Address, 0),
		Sender:    make([]*Address, 0),
		ReplyTo:   make([]*Address, 0),
		To:        make([]*Address, 0),
		CC:        make([]*Address, 0),
		BCC:       make([]*Address, 0),
		InReplyTo: make([]*Address, 0),
	}
}

func (e *Envelope) SetDate(date string) *Envelope {
	e.Date = date

	return e
}

func (e *Envelope) SetSubject(sub string) *Envelope {
	e.Subject = sub

	return e
}

func (e *Envelope) AddFromAddr(addr *Address) *Envelope {
	e.From = append(e.From, addr)

	return e
}

func (e *Envelope) AddSenderAddr(addr *Address) *Envelope {
	e.Sender = append(e.Sender, addr)

	return e
}

func (e *Envelope) AddReplyToAddr(addr *Address) *Envelope {
	e.ReplyTo = append(e.ReplyTo, addr)

	return e
}

func (e *Envelope) AddToAddr(addr *Address) *Envelope {
	e.To = append(e.To, addr)

	return e
}

func (e *Envelope) AddCCAddr(addr *Address) *Envelope {
	e.CC = append(e.CC, addr)

	return e
}

func (e *Envelope) AddBCCAddr(addr *Address) *Envelope {
	e.BCC = append(e.BCC, addr)

	return e
}

func (e *Envelope) AddInReplyToAddr(addr *Address) *Envelope {
	e.InReplyTo = append(e.InReplyTo, addr)

	return e
}

func (e *Envelope) SetMessageID(msgID string) *Envelope {
	e.MessageID = msgID

	return e
}
