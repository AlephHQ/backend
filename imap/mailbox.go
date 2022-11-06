package imap

type Mailbox struct {
	Name string

	Flags          []string
	PermanentFlags []string

	Exists uint64

	Recent uint64

	Unseen uint64

	UIDValidity uint64

	UIDNext uint64

	ReadOnly bool
}

func NewMailbox() *Mailbox {
	return &Mailbox{}
}

func (mbs *Mailbox) SetName(name string) *Mailbox {
	mbs.Name = name

	return mbs
}

func (mbs *Mailbox) SetFlags(flags []string) *Mailbox {
	mbs.Flags = make([]string, 0)
	mbs.Flags = append(mbs.Flags, flags...)

	return mbs
}

func (mbs *Mailbox) SetPermanentFlags(pflags []string) *Mailbox {
	mbs.PermanentFlags = make([]string, 0)
	mbs.PermanentFlags = append(mbs.PermanentFlags, pflags...)

	return mbs
}

func (mbs *Mailbox) SetExists(exists uint64) *Mailbox {
	mbs.Exists = exists

	return mbs
}

func (mbs *Mailbox) SetRecent(recent uint64) *Mailbox {
	mbs.Recent = recent

	return mbs
}

func (mbs *Mailbox) SetUnseen(unseen uint64) *Mailbox {
	mbs.Unseen = unseen

	return mbs
}

func (mbs *Mailbox) SetUIDValidity(uidvalidity uint64) *Mailbox {
	mbs.UIDValidity = uidvalidity

	return mbs
}

func (mbs *Mailbox) SetUIDNext(uidnext uint64) *Mailbox {
	mbs.UIDNext = uidnext

	return mbs
}

func (mbs *Mailbox) SetReadOnly(readonly bool) *Mailbox {
	mbs.ReadOnly = readonly

	return mbs
}
