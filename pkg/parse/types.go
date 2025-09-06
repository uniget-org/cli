package parse

import "github.com/regclient/regclient/types/ref"

type ImageRefs struct {
	Refs []ref.Ref
}

func (r *ImageRefs) Add(ref ref.Ref) {
	r.Refs = append(r.Refs, ref)
}
