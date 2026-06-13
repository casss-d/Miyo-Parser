package miyo

type Possibility struct {
	Tag        Tag
	Descriptor Tag
	Score      float64
	Value      string
	Source     string
}

func (p Possibility) effectiveDescriptor() Tag {
	if p.Descriptor != "" {
		return p.Descriptor
	}
	return p.Tag
}
