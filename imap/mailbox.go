package imap

type Mailbox struct {
	Name string `json:"name"`

	Flags          map[Flag]bool `json:"flags"`
	PermanentFlags map[Flag]bool `json:"permament_flags"`

	// Total number of messages in mailbox
	Exists uint64 `json:"exists"`

	// Number of recent messages
	Recent uint64 `json:"recent"`

	// Sequence number of the first unseen message
	Unseen uint64 `json:"unseen"`

	UIDValidity uint64 `json:"uid_validity"`

	UIDNext uint64 `json:"uid_next"`

	ReadOnly bool `json:"read_only"`
}

func NewMailbox() *Mailbox {
	return &Mailbox{}
}

func (mbs *Mailbox) SetName(name string) *Mailbox {
	mbs.Name = name

	return mbs
}

func (mbs *Mailbox) SetFlags(flags []string) *Mailbox {
	mbs.Flags = make(map[Flag]bool)
	for _, f := range flags {
		mbs.Flags[Flag(f)] = true
	}

	return mbs
}

func (mbs *Mailbox) SetPermanentFlags(pflags []string) *Mailbox {
	mbs.PermanentFlags = make(map[Flag]bool)
	for _, f := range pflags {
		mbs.PermanentFlags[Flag(f)] = true
	}

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
