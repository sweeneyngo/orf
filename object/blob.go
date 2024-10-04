package object

type Blob struct {
	Object
}

func CreateBlob(data []byte) *Blob {
	return &Blob{
		Object: Object{
			format: "blob",            // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			Data:   data,              // Set the data directly in the Object
		},
	}
}

func (blob *Blob) Serialize() []byte {
	return blob.Data
}

func (blob *Blob) Deserialize(data []byte) {
	blob.Data = data
}
