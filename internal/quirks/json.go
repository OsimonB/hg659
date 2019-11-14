package quirks

import (
	"bytes"
	"io"
	"io/ioutil"
)

// ReadAll extracts the contents between /* and */. If delimiters are
// not found, it behaves the same as ioutil.ReadAll.
func ReadAll(r io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	start := bytes.Index(data, []byte("/*"))
	end := bytes.Index(data, []byte("*/"))

	if start == -1 || end == -1 {
		return data, nil
	}

	return data[start+2 : end], nil
}
