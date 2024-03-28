package label

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"go.mongodb.org/mongo-driver/bson"
)

type FilterOptions struct {
	OnlyZone    *string
	OnlySubZone *string
}

type MongoIndexFilter struct {
	bson.M
	*FilterOptions
}

func makeMongoFilter(opts *FilterOptions) bson.M {
	return (&MongoIndexFilter{
		M:             bson.M{},
		FilterOptions: opts,
	}).toBson()
}

func (f *MongoIndexFilter) toBson() bson.M {
	f.CheckOnlyZone()
	f.CheckOnlySubZone()
	return f.M
}

func (f *MongoIndexFilter) CheckOnlyZone() {
	if f.OnlyZone != nil {
		f.M[consts.Zone] = *f.OnlyZone
	}
}

func (f *MongoIndexFilter) CheckOnlySubZone() {
	if f.OnlySubZone != nil {
		f.M[consts.SubZone] = *f.OnlySubZone
	}
}
