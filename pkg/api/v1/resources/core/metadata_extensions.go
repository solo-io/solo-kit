package core

func (m *Metadata) Less(m2 *Metadata) bool {
	if m.GetNamespace() == m2.GetNamespace() {
		return m.GetName() < m2.GetName()
	}
	return m.GetNamespace() < m2.GetNamespace()
}

func (m *Metadata) Ref() *ResourceRef {
	return &ResourceRef{
		Namespace: m.GetNamespace(),
		Name:      m.GetName(),
	}
}

func (m *Metadata) Match(ref *ResourceRef) bool {
	return m.GetNamespace() == ref.GetNamespace() && m.GetName() == ref.GetName()
}

type Predicate func(metadata *Metadata) bool
