package httpevents

import (
	"fmt"

	"github.com/segmentio/objconv"
)

// An ordered list of HTTP headers.
type headerList []header

type header struct {
	name  string
	value string
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
