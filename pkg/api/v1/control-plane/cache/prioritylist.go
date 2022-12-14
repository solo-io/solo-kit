package cache

import (
	"fmt"
	"sort"
)

// The default Priority is the last index. The last index is not required in the priority list, thus any elements added
// to the struct will be added to the default priority (last).
const defaultPriority = -1

type PriorityValue interface {
	GetPriority() string
}

// PrioritySortedStruct inserts elements into a map indexed in a list by prioirity. Priority lists are unordered.
// This allows O(1) inserts, gets, and deletes of these elements as long as the client has the index of the element.
// Indexes are returned on Add() function, when elements are added to the collection.
//
// Elements interface the PriorityValue interface, this allows the collection to call the element to get it's priority.
// If a priority does not exist for an added element, it is added to the lowest priority.
//
// The below 1, 2, 4 go first, 10 second, and 17 third. Anything else is processed last
//
//	{
//		0: {item1, item2, item4}
//		1: {item10}
//		2: {item17}
//	}
type PrioritySortedStruct struct {
	// priorityMap maps the value to the priority.
	priorityMap map[string]int
	// elements are the map of elements structured by their priority.
	//
	// This is a list of maps with the key being an int. Because we do not care about order, and want to maintain
	// O(1) delete, get, and insert we maintain the elements in maps with an index that gets incremented on every insert.
	// This way clients can get and insert the data with ids generated by the collection itself.
	elements []map[uint64]ResponseWatch
	// nextUniqueElementIndex is the next element index to be inserted.
	//
	// We have a list of maps we need to ensure that when adding to the struct that the
	// elements do not replace an element previously inserted. To maintain unique indexes within
	// the map we have to keep a running index of all the number of elements inserted into the
	// collection.
	nextUniqueElementIndex uint64
}

// PriorityIndex is the priority and index used to locate items.
type PriorityIndex struct {
	Priority int
	Index    uint64
}

// NewPrioritySortedStruct creates a new Priority Sorted Struct.
// prioritySets is the set lists for priorities, where P is the type used for priority.
func NewPrioritySortedStruct(prioritySets map[int][]string) *PrioritySortedStruct {
	// need to ensure that the prioriries are in order and there are no missing or skipped Priorities
	priorities := make([]int, 0)
	for priority := range prioritySets {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)
	currentP := 0
	for _, p := range priorities {
		if currentP == p {
			currentP++
		} else {
			panic(fmt.Sprintf("Priorities are not set correct, you are missing priority %d", currentP))
		}
	}
	// +1 for last priority list
	elements := make([]map[uint64]ResponseWatch, len(prioritySets)+1)
	if len(prioritySets) > 0 {
		for priorityIndex := range prioritySets {
			elements[priorityIndex] = make(map[uint64]ResponseWatch)
		}
		elements[len(prioritySets)] = make(map[uint64]ResponseWatch)
	} else {
		// there is only one map of elements
		elements[0] = make(map[uint64]ResponseWatch)
	}
	priorityMap := make(map[string]int)
	for index, pl := range prioritySets {
		for _, v := range pl {
			priorityMap[v] = index
		}
	}
	p := PrioritySortedStruct{
		elements:    elements,
		priorityMap: priorityMap,
	}
	return &p
}

// Get returns the element at the index, and if it exists.
func (p *PrioritySortedStruct) Get(pi PriorityIndex) (ResponseWatch, bool) {
	v, ok := p.elements[pi.Priority][pi.Index]
	return v, ok
}

// Process will call the procesFunc over all the elements by priority.
func (p *PrioritySortedStruct) Process(processFunc func(el ResponseWatch, pi PriorityIndex)) {
	for i := 0; i < len(p.elements); i++ {
		m := p.elements[i]
		for index, v := range m {
			processFunc(v, PriorityIndex{Priority: i, Index: index})
		}
	}
}

// GetPriorityList returns an ordered list of the elements by priority.
func (p *PrioritySortedStruct) GetPriorityList() []ResponseWatch {
	elements := make([]ResponseWatch, 0, p.Len())
	for priority := 0; priority < len(p.elements); priority++ {
		mapOfElements := p.elements[priority]
		for _, el := range mapOfElements {
			elements = append(elements, el)
		}
	}
	return elements
}

// Add will add the element to the Priority Collection, returns the priority, and element number.
func (p *PrioritySortedStruct) Add(element ResponseWatch) PriorityIndex {
	priority := p.getPriorityOfElement(element)
	if priority == defaultPriority {
		// add to the last index of the watches
		priority = len(p.elements) - 1
	}
	p.elements[priority][p.nextUniqueElementIndex] = element
	pi := PriorityIndex{Priority: priority, Index: p.nextUniqueElementIndex}
	p.nextUniqueElementIndex++
	return pi
}

// Delete will delete the element, returns true if it deleted.
func (p *PrioritySortedStruct) Delete(pi PriorityIndex) bool {
	if p.Len() == 0 {
		return false
	}
	if _, ok := p.Get(pi); ok {
		delete(p.elements[pi.Priority], pi.Index)
		return ok
	} else {
		return false
	}
}

// Len will return the number of elements
func (p *PrioritySortedStruct) Len() int {
	count := 0
	for _, el := range p.elements {
		count += len(el)
	}
	return count
}

// GetPriorityIndexes returns a list of all the indexes for all elements by priority.
func (p *PrioritySortedStruct) GetPriorityIndexes() []PriorityIndex {
	pi := make([]PriorityIndex, 0, p.Len())
	for i := 0; i < len(p.elements); i++ {
		m := p.elements[i]
		for index := range m {
			pi = append(pi, PriorityIndex{Priority: i, Index: index})
		}
	}
	return pi
}

// getPriorityOfElement returns the priority of element ResponseWatch.
func (p *PrioritySortedStruct) getPriorityOfElement(element ResponseWatch) int {
	pv := element.GetPriority()
	if p, exists := p.priorityMap[pv]; exists {
		return p
	} else {
		// default priority is -1
		return defaultPriority
	}
}
