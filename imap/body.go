package imap

type Body struct {
	Parts            []*BodyStructure
	Multipart        bool
	MultipartSubtype string
	Sections         map[string]string
	Full             string
}

func NewBody() *Body {
	return &Body{
		Parts:    make([]*BodyStructure, 0),
		Sections: make(map[string]string),
	}
}

func (b *Body) AddPart(p *BodyStructure) *Body {
	b.Parts = append(b.Parts, p)

	return b
}

func (b *Body) SetMultipart(multi bool) *Body {
	b.Multipart = multi

	return b
}

func (b *Body) SetMultipartSubtype(st string) *Body {
	b.MultipartSubtype = st

	return b
}

func (b *Body) SetSection(key, val string) *Body {
	b.Sections[key] = val

	return b
}
