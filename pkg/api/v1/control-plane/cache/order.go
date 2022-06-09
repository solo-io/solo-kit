package cache

// Key is an internal sorting data structure we can use to
// order responses by Type and their associated watch IDs.
type key struct {
	ID      int64
	TypeURL string
}

type orderingList struct {
	keys []key
	// typeWeights for string o(1) lookup
	typeWeights map[string]int
}

func (ol orderingList) Len() int {
	return len(ol.keys)
}

//-- need to create an ordering similar to that of upstream --
// Less compares the typeURL and determines what order things should be sent.
func (ol orderingList) Less(i, j int) bool {
	return ol.typeWeights[ol.keys[i].TypeURL] > ol.typeWeights[ol.keys[j].TypeURL]
}

func (ol orderingList) Swap(i, j int) {
	ol.keys[i], ol.keys[j] = ol.keys[j], ol.keys[i]
}
