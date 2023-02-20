package api

type Items struct {
	items    []string
	position int
}

func (i *Items) ValidPosition() bool {
	return i.position < len(i.items)
}

func (i *Items) Item() string {
	if i.ValidPosition() {
		return i.items[i.position]
	}
	panic("Error: Outside of valid position")
}

func (i *Items) Next() {
	i.position++
}

func (i *Items) NextItem() string {
	i.Next()
	return i.Item()
}

func (i *Items) PeekItem() string {
	if i.position+1 < len(i.items) {
		return i.items[i.position+1]
	}
	panic("Error: Peeked outside of Valid Position")
}

func InitItems(items []string) Items {
	return Items{
		items:    items,
		position: 0,
	}
}
