package labelEntity

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"go.mongodb.org/mongo-driver/bson"
)

type FilterOptions struct {
	OnlyLabelId    *string
	OnlyObjectType *int64
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
	f.CheckOnlyLabelId()
	f.CheckOnlyObjectType()
	return f.M
}

func (f *MongoIndexFilter) CheckOnlyLabelId() {
	if f.OnlyLabelId != nil {
		f.M[consts.Labels] = *f.OnlyLabelId
	}
}

func (f *MongoIndexFilter) CheckOnlyObjectType() {
	if f.OnlyObjectType != nil {
		f.M[consts.ObjectType] = *f.OnlyObjectType
	}
}
