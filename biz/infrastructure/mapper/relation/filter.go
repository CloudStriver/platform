package relation

import "github.com/CloudStriver/platform-comment/biz/infrastructure/consts"

type FilterOptions struct {
	OnlyFromType     *int64
	OnlyFromId       *string
	OnlyToType       *int64
	OnlyToId         *string
	OnlyRelationType *int64
}

type Neo4jFilter struct {
	m map[string]any
	*FilterOptions
}

func (f *Neo4jFilter) CheckOnlyToId() {
	if f.OnlyToId != nil {
		f.m[consts.ToId] = *f.OnlyToId
	}
}

func (f *Neo4jFilter) CheckOnlyToType() {
	if f.OnlyToType != nil {
		f.m[consts.ToType] = *f.OnlyToType
	}
}

func (f *Neo4jFilter) CheckOnlyFromId() {
	if f.OnlyFromId != nil {
		f.m[consts.FromId] = *f.OnlyFromId
	}
}

func (f *Neo4jFilter) CheckOnlyFromType() {
	if f.OnlyFromType != nil {
		f.m[consts.FromType] = *f.OnlyFromType
	}
}

func (f *Neo4jFilter) CheckOnlyRelationType() {
	if f.OnlyRelationType != nil {
		f.m[consts.RelationType] = *f.OnlyRelationType
	}
}
