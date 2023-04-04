package httpevents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// An ordered list of HTTP headers.
type headerList []header

type header struct {
	name  string
	value string
}

func (h headerList) clear() {
	for i := range h {
		h[i] = header{}
	}
}

func (h *headerList) set(httpHeader http.Header) {
	list := (*h)[:0]

	for name, values := range httpHeader {
		// Do not accept and log headers that contain credentials
		if name == "Authorization" {
			continue
		}

		for _, value := range values {
			list = append(list, header{
				name:  name,
				value: value,
			})
		}
	}

	*h = list
	sort.Sort(h)
}

func (h *headerList) String() string {
	return fmt.Sprint(*h)
}

func (h *headerList) Clone() interface{} {
	c := make(headerList, len(*h))
	copy(c, *h)
	return &c
}

func (h *headerList) MarshalJSON() ([]byte, error) {
	if len(*h) == 0 {
		return []byte(`{}`), nil
	}

	b := &bytes.Buffer{}
	b.Grow(256)
	b.WriteByte('{')

	for i := range *h {
		if i != 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(b, `%q:%q`, jsonString{&(*h)[i].name}, jsonString{&(*h)[i].value})
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

func (h *headerList) Len() int               { return len(*h) }
func (h *headerList) Less(i int, j int) bool { return (*h)[i].name < (*h)[j].name }
func (h *headerList) Swap(i int, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

type jsonString struct{ s *string }

func (j jsonString) String() string { return *j.s }

var _ json.Marshaler = (*headerList)(nil)
