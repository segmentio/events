package httpevents

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/segmentio/objconv"
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

func (h *headerList) EncodeValue(e objconv.Encoder) error {
	i := 0
	return e.EncodeMap(len(*h), func(k objconv.Encoder, v objconv.Encoder) error {
		if err := k.Encode(&(*h)[i].name); err != nil {
			return err
		}
		if err := v.Encode(&(*h)[i].value); err != nil {
			return err
		}
		i++
		return nil
	})
}

func (h *headerList) Len() int               { return len(*h) }
func (h *headerList) Less(i int, j int) bool { return (*h)[i].name < (*h)[j].name }
func (h *headerList) Swap(i int, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
