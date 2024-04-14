package comment

import (
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilterOptions struct {
	OnlyUserId     *string
	OnlyAtUserId   *string
	OnlyRootId     *string
	OnlyCommentIds []string
	OnlyState      *int64
	OnlyAttrs      *int64
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
	f.CheckOnlyUserId()
	f.CheckOnlyAtUserId()
	f.CheckOnlyCommentIds()
	f.CheckOnlyRootId()
	f.CheckOnlyState()
	f.CheckOnlyAttrs()
	return f.m
}

func (f *MongoFilter) CheckOnlyCommentIds() {
	if f.OnlyCommentIds != nil {
		f.m[consts.ID] = bson.M{
			"$in": lo.Map[string, primitive.ObjectID](f.OnlyCommentIds, func(s string, _ int) primitive.ObjectID {
				oid, _ := primitive.ObjectIDFromHex(s)
				return oid
			}),
		}
	}
}

func (f *MongoFilter) CheckOnlyUserId() {
	if f.OnlyUserId != nil {
		f.m[consts.UserId] = *f.OnlyUserId
	}
}

func (f *MongoFilter) CheckOnlyRootId() {
	if f.OnlyRootId != nil {
		f.m[consts.RootId] = *f.OnlyRootId
	}
}

func (f *MongoFilter) CheckOnlyAtUserId() {
	if f.OnlyAtUserId != nil {
		f.m[consts.AtUserId] = *f.OnlyAtUserId
	}
}

func (f *MongoFilter) CheckOnlyState() {
	if f.OnlyState != nil {
		f.m[consts.State] = *f.OnlyState
	}
}

func (f *MongoFilter) CheckOnlyAttrs() {
	if f.OnlyAttrs != nil {
		f.m[consts.Attrs] = *f.OnlyAttrs
	}
}
