package object

type Blob struct {
	Base
}

func (blob *Blob) GetFormat() string {
	return blob.format
}

func (blob *Blob) GetSize() uint32 {
	return blob.size
}

func (blob *Blob) GetData() []byte {
	return blob.data
}

// Creates a new Blob object.
func CreateBlob(data []byte) *Blob {
	return &Blob{
		Base: Base{
			format: "blob",            // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			data:   data,              // Set the data directly in the Object
		},
	}
}

func (blob *Blob) Serialize() []byte {
	return blob.data
}

func (blob *Blob) Deserialize(data []byte) {
	blob.data = data
}
