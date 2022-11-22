package imap

type DataItemName string

const (
	DataItemNameBody         DataItemName = "BODY"
	DataItemNameBodyPeek     DataItemName = "BODY.PEEK"
	DataItemNameInternalDate DataItemName = "INTERNALDATE"
	DataItemNameRFC822Size   DataItemName = "RFC822.SIZE"
	DataItemNameEnvelope     DataItemName = "ENVELOPE"
	DataItemNameRFC822Header DataItemName = "RFC822.HEADER"
	DataItemNameUID          DataItemName = "UID"
	DataItemNameFlags        DataItemName = "FLAGS"
	DataItemNameRFC822       DataItemName = "RFC822"

	DataItemPlusFlag  DataItemName = "+FLAGS"
	DataItemMinusFlag DataItemName = "-FLAGS"
)

type BodySection string

const (
	BodySectionText   BodySection = "TEXT"
	BodySectionHeader BodySection = "HEADER"
)

type DataItem struct {
	Name    DataItemName
	Section BodySection
	Partial string
}

func NewDataItem(name DataItemName) *DataItem {
	return &DataItem{Name: name}
}

func (di *DataItem) SetSection(section BodySection) *DataItem {
	di.Section = section

	return di
}
