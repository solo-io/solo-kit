package core

import (
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
)

// collapse a status into a status with no children
func (s Status) Flatten() Status {
	if len(s.SubresourceStatuses) == 0 {
		return s
	}
	out := Status{
		State:  s.State,
		Reason: s.Reason,
	}
	orderedSubStatusMapIterator(s.SubresourceStatuses, func(key string, stat *Status_SubStatus) {
		switch stat.State {
		case Status_Rejected:
			out.State = Status_Rejected
			out.Reason += key + fmt.Sprintf("child %v rejected with reason: %v.\n", key, stat.Reason)
		case Status_Pending:
			if out.State == Status_Accepted {
				out.State = Status_Pending
			}
			out.Reason += key + " is still pending.\n"
		}
	})
	return out
}
func (s Status) DeepCopyInto(out *Status) {
	clone := proto.Clone(&s).(*Status)
	*out = *clone
}

func orderedMapIterator(m map[string]*Status, onKey func(key string, value *Status)) {
	var list []struct {
		key   string
		value *Status
	}
	for k, v := range m {
		list = append(list, struct {
			key   string
			value *Status
		}{
			key:   k,
			value: v,
		})
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].key < list[j].key
	})
	for _, el := range list {
		onKey(el.key, el.value)
	}
}

func orderedSubStatusMapIterator(m map[string]*Status_SubStatus, onKey func(key string, value *Status_SubStatus)) {
	var list []struct {
		key   string
		value *Status_SubStatus
	}
	for k, v := range m {
		list = append(list, struct {
			key   string
			value *Status_SubStatus
		}{
			key:   k,
			value: v,
		})
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].key < list[j].key
	})
	for _, el := range list {
		onKey(el.key, el.value)
	}
}
