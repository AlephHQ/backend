package imap

type MailboxStatus struct {
	Name string

	Flags          []string
	PermanentFlags []string

	Exists uint

	Recent uint

	Unseen uint

	UIDValidity uint

	UIDNext uint

	ReadOnly bool
}

func NewMailboxStatus() *MailboxStatus {
	return &MailboxStatus{}
}

func (mbs *MailboxStatus) SetName(name string) *MailboxStatus {
	mbs.Name = name

	return mbs
}

func (mbs *MailboxStatus) SetFlags(flags []string) *MailboxStatus {
	mbs.Flags = make([]string, 0)
	mbs.Flags = append(mbs.Flags, flags...)

	return mbs
}

func (mbs *MailboxStatus) SetPermanentFlags(pflags []string) *MailboxStatus {
	mbs.PermanentFlags = make([]string, 0)
	mbs.PermanentFlags = append(mbs.PermanentFlags, pflags...)

	return mbs
}

func (mbs *MailboxStatus) SetExists(exists uint) *MailboxStatus {
	mbs.Exists = exists

	return mbs
}

func (mbs *MailboxStatus) SetRecent(recent uint) *MailboxStatus {
	mbs.Recent = recent

	return mbs
}

func (mbs *MailboxStatus) SetUnseen(unseen uint) *MailboxStatus {
	mbs.Unseen = unseen

	return mbs
}

func (mbs *MailboxStatus) SetUIDValidity(uidvalidity uint) *MailboxStatus {
	mbs.UIDValidity = uidvalidity

	return mbs
}

func (mbs *MailboxStatus) SetUIDNext(uidnext uint) *MailboxStatus {
	mbs.UIDNext = uidnext

	return mbs
}

func (mbs *MailboxStatus) SetReadOnly(readonly bool) *MailboxStatus {
	mbs.ReadOnly = readonly

	return mbs
}
