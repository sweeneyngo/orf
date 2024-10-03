package object

// Format for the blob object
const format = "blob"

type Blob struct {
	data []byte
}

func CreateBlob(data []byte) *Blob {
	return &Blob{
		data: data,
	}
}

func (blob *Blob) Serialize() []byte {
	return blob.data
}

func (blob *Blob) Deserialize(data []byte) {
	blob.data = data
}
