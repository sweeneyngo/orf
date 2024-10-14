package object

import "orf/kv"

// Represents a tag object, with key-value data forming the tag message.
type Tag struct {
	Base
	kvData *kv.OrderedMap
}

func (tag *Tag) GetFormat() string {
	return tag.format
}

func (tag *Tag) GetSize() uint32 {
	return tag.size
}

func (tag *Tag) GetData() []byte {
	return tag.data
}

func (tag *Tag) GetKVData() *kv.OrderedMap {
	return tag.kvData
}

// Creates a new tag object with empty key-value data.
func Createtag(data []byte) *Tag {
	return &Tag{
		Base: Base{
			format: "tag",             // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			data:   data,              // Set the data directly in the Object
		},
		kvData: nil,
	}
}

func (tag *Tag) Serialize() []byte {
	return kv.Serialize(tag.kvData)
}

func (tag *Tag) Deserialize(data []byte) error {
	kvData, err := kv.Parse(data, 0, tag.kvData)
	if err != nil {
		return err
	}

	tag.kvData = kvData
	return nil
}
