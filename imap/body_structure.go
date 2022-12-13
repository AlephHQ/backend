package imap

type BodyStructure struct {
	Type          string
	Subtype       string
	ParameterList map[string]string
	ID            string
	Description   string
	Encoding      string
	Size          uint64
	SizeInLines   uint64
}

func NewBodyStrcuture() *BodyStructure {
	return &BodyStructure{
		ParameterList: make(map[string]string),
	}
}

func (bs *BodyStructure) SetType(t string) *BodyStructure {
	bs.Type = t

	return bs
}

func (bs *BodyStructure) SetSubtype(st string) *BodyStructure {
	bs.Subtype = st

	return bs
}

func (bs *BodyStructure) AddKeyValParam(key, val string) *BodyStructure {
	bs.ParameterList[key] = val

	return bs
}

func (bs *BodyStructure) SetID(id string) *BodyStructure {
	bs.ID = id

	return bs
}

func (bs *BodyStructure) SetDescription(desc string) *BodyStructure {
	bs.Description = desc

	return bs
}

func (bs *BodyStructure) SetEncoding(enc string) *BodyStructure {
	bs.Encoding = enc

	return bs
}

func (bs *BodyStructure) SetSize(size uint64) *BodyStructure {
	bs.Size = size

	return bs
}

func (bs *BodyStructure) SetSizeInLines(lines uint64) *BodyStructure {
	bs.SizeInLines = lines

	return bs
}
