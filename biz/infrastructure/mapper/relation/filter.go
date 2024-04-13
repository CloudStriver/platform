package relation

import (
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"go.mongodb.org/mongo-driver/bson"
)

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

type MongoFilter struct {
	m bson.M
	*FilterOptions
}

func makeMongoFilter(opts *FilterOptions) bson.M {
	return (&MongoFilter{
		m:             bson.M{},
		FilterOptions: opts,
	}).toBson()
}

func (f *MongoFilter) toBson() bson.M {
	f.CheckOnlyFromId()
	f.CheckOnlyFromType()
	f.CheckOnlyToId()
	f.CheckOnlyToType()
	f.CheckOnlyRelationType()
	return f.m
}

func (f *MongoFilter) CheckOnlyToId() {
	if f.OnlyToId != nil {
		f.m[consts.ToId] = *f.OnlyToId
	}
}

func (f *MongoFilter) CheckOnlyToType() {
	if f.OnlyToType != nil {
		f.m[consts.ToType] = *f.OnlyToType
	}
}

func (f *MongoFilter) CheckOnlyFromId() {
	if f.OnlyFromId != nil {
		f.m[consts.FromId] = *f.OnlyFromId
	}
}

func (f *MongoFilter) CheckOnlyFromType() {
	if f.OnlyFromType != nil {
		f.m[consts.FromType] = *f.OnlyFromType
	}
}

func (f *MongoFilter) CheckOnlyRelationType() {
	if f.OnlyRelationType != nil {
		f.m[consts.RelationType] = *f.OnlyRelationType
	}
}
