package subject

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilterOptions struct {
	OnlyUserId    *string
	OnlyItemId    *string
	OnlySubjectId *string
	OnlyState     *int64
	OnlyAttrs     *int64
}

type MongoSubjectFilter struct {
	bson.M
	*FilterOptions
}

func makeMongoFilter(opts *FilterOptions) bson.M {
	return (&MongoSubjectFilter{
		M:             bson.M{},
		FilterOptions: opts,
	}).toBson()
}

func (f *MongoSubjectFilter) toBson() bson.M {
	f.CheckOnlyUserId()
	f.CheckOnlyItemId()
	f.CheckOnlySubjectId()
	f.CheckOnlyState()
	f.CheckOnlyAttrs()
	return f.M
}

func (f *MongoSubjectFilter) CheckOnlyUserId() {
	if f.OnlyUserId != nil {
		f.M[consts.UserId] = *f.OnlyUserId
	}
}

func (f *MongoSubjectFilter) CheckOnlyItemId() {
	if f.OnlyItemId != nil {
		f.M[consts.ItemId] = *f.OnlyItemId
	}
}

func (f *MongoSubjectFilter) CheckOnlySubjectId() {
	if f.OnlySubjectId != nil {
		oid, _ := primitive.ObjectIDFromHex(*f.OnlySubjectId)
		f.M[consts.ID] = oid
	}
}

func (f *MongoSubjectFilter) CheckOnlyState() {
	if f.OnlyState != nil {
		f.M[consts.State] = *f.OnlyState
	}
}

func (f *MongoSubjectFilter) CheckOnlyAttrs() {
	if f.OnlyAttrs != nil {
		f.M[consts.Attrs] = *f.OnlyAttrs
	}
}
