package stats

import "go.opencensus.io/tag"

var (
	NamespaceKey  = MustTag("namespace")
	ResourceKey   = MustTag("resource")
	SyncerNameKey = MustTag("syncer_name")
)

func MustTag(name string) tag.Key {
	t, err := tag.NewKey(name)
	if err != nil {
		panic(err)
	}
	return t
}
