package comment

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilterOptions struct {
	OnlyUserId     *string
	OnlyAtUserId   *string
	OnlyCommentId  *string
	OnlySubjectId  *string
	OnlyRootId     *string
	OnlyFatherId   *string
	OnlyCommentIds []string
	OnlyState      *int64
	OnlyAttrs      *int64
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
	f.CheckOnlyUserId()
	f.CheckOnlyAtUserId()
	f.CheckOnlyCommentId()
	f.CheckOnlySubjectId()
	f.CheckOnlyRootId()
	f.CheckOnlyFatherId()
	f.CheckOnlyState()
	f.CheckOnlyAttrs()
	return f.M
}

func (f *MongoIndexFilter) CheckOnlyFileIds() {
	if f.OnlyCommentIds != nil {
		f.M[consts.ID] = bson.M{
			"$in": lo.Map[string, primitive.ObjectID](f.OnlyCommentIds, func(s string, _ int) primitive.ObjectID {
				oid, _ := primitive.ObjectIDFromHex(s)
				return oid
			}),
		}
	}
}

func (f *MongoIndexFilter) CheckOnlyUserId() {
	if f.OnlyUserId != nil {
		f.M[consts.UserId] = *f.OnlyUserId
	}
}

func (f *MongoIndexFilter) CheckOnlyAtUserId() {
	if f.OnlyAtUserId != nil {
		f.M[consts.AtUserId] = *f.OnlyAtUserId
	}
}

func (f *MongoIndexFilter) CheckOnlyCommentId() {
	if f.OnlyCommentId != nil {
		oid, _ := primitive.ObjectIDFromHex(*f.OnlyCommentId)
		f.M[consts.ID] = oid
	}
}

func (f *MongoIndexFilter) CheckOnlySubjectId() {
	if f.OnlySubjectId != nil {
		f.M[consts.SubjectId] = *f.OnlySubjectId
	}
}

func (f *MongoIndexFilter) CheckOnlyRootId() {
	if f.OnlyRootId != nil {
		f.M[consts.RootId] = *f.OnlyRootId
	}
}

func (f *MongoIndexFilter) CheckOnlyFatherId() {
	if f.OnlyFatherId != nil {
		f.M[consts.FatherId] = *f.OnlyFatherId
	}
}

func (f *MongoIndexFilter) CheckOnlyState() {
	if f.OnlyState != nil {
		f.M[consts.State] = *f.OnlyState
	}
}

func (f *MongoIndexFilter) CheckOnlyAttrs() {
	if f.OnlyAttrs != nil {
		f.M[consts.Attrs] = *f.OnlyAttrs
	}
}
