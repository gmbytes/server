package data

import "server/data/enum"

type Attr struct {
	Type enum.AttrType
	Val  int64
	Rate int64
}

type Attrs []*Attr

func (ss *Attrs) GetValue(ty enum.AttrType) int64 {
	for _, attr := range *ss {
		if attr.Type == ty {
			return attr.Val
		}
	}
	return 0
}
