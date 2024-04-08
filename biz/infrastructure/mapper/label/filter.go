package label

import (
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"go.mongodb.org/mongo-driver/bson"
)

type FilterOptions struct {
	OnlyFatherId *string
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
	f.CheckOnlyFatherId()
	return f.M
}

func (f *MongoIndexFilter) CheckOnlyFatherId() {
	if f.OnlyFatherId != nil {
		f.M[consts.FatherId] = *f.OnlyFatherId
	}
}
