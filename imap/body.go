package imap

type Body struct {
	Parts            []*BodyStructure
	Multipart        bool
	MultipartSubtype string
}

func NewBody() *Body {
	return &Body{
		Parts: make([]*BodyStructure, 0),
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
