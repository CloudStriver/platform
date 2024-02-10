package convertor

import (
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/comment"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/label"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/labelEntity"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/subject"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/basic"
	gencomment "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CheckId(id string) primitive.ObjectID {
	var oid primitive.ObjectID
	if id == "" {
		oid = primitive.NilObjectID
	} else {
		oid, _ = primitive.ObjectIDFromHex(id)
	}
	return oid
}

func CommentToCommentMapper(data *gencomment.Comment) *comment.Comment {
	return &comment.Comment{
		ID:        CheckId(data.Id),
		UserId:    data.UserId,
		AtUserId:  data.AtUserId,
		SubjectId: data.SubjectId,
		RootId:    data.RootId,
		FatherId:  data.FatherId,
		Content:   data.Content,
		Meta:      data.Meta,
		Tags:      data.Tags,
		Count:     data.Count,
		State:     int64(data.State),
		Attrs:     int64(data.Attrs),
		SortTime:  data.CreateTime,
	}
}

func CommentMapperToCommentInfo(data *comment.Comment) *gencomment.CommentInfo {
	return &gencomment.CommentInfo{
		Id:         data.ID.Hex(),
		SubjectId:  data.SubjectId,
		RootId:     data.RootId,
		FatherId:   data.FatherId,
		Count:      *data.Count,
		State:      data.State,
		Attrs:      data.Attrs,
		Tags:       data.Tags,
		UserId:     data.UserId,
		AtUserId:   data.AtUserId,
		Content:    data.Content,
		Meta:       data.Meta,
		CreateTime: data.CreateAt.UnixMilli(),
	}
}

func CommentFilterOptionsToFilterOptions(data *gencomment.CommentFilterOptions) *comment.FilterOptions {
	if data == nil {
		return &comment.FilterOptions{}
	} else {
		return &comment.FilterOptions{
			OnlyUserId:    data.OnlyUserId,
			OnlyAtUserId:  data.OnlyAtUserId,
			OnlyCommentId: data.OnlyCommentId,
			OnlySubjectId: data.OnlySubjectId,
			OnlyRootId:    data.OnlyRootId,
			OnlyFatherId:  data.OnlyFatherId,
			OnlyState:     data.OnlyState,
			OnlyAttrs:     data.OnlyAttrs,
		}
	}
}

func SubjectFilterOptionsToFilterOptions(data *gencomment.SubjectFilterOptions) *subject.FilterOptions {
	if data == nil {
		return &subject.FilterOptions{}
	} else {
		return &subject.FilterOptions{
			OnlyUserId:    data.OnlyUserId,
			OnlyItemId:    data.OnlyItemId,
			OnlySubjectId: data.OnlySubjectId,
			OnlyState:     data.OnlyState,
			OnlyAttrs:     data.OnlyAttrs,
		}
	}
}

func SubjectToSubjectMapper(data *gencomment.Subject) *subject.Subject {
	return &subject.Subject{
		ID:           CheckId(data.Id),
		ItemId:       data.ItemId,
		UserId:       data.UserId,
		TopCommentId: lo.ToPtr(data.TopCommentId),
		RootCount:    data.RootCount,
		AllCount:     data.AllCount,
		State:        int64(data.State),
		Attrs:        int64(data.Attrs),
	}
}

func SubjectMapperToSubjectInfo(data *subject.Subject) *gencomment.SubjectInfo {
	return &gencomment.SubjectInfo{
		Id:     data.ID.Hex(),
		ItemId: data.ItemId,
		UserId: data.UserId,
		Attrs:  data.Attrs,
	}
}

func SubjectMapperToSubjectDetail(data *subject.Subject) *gencomment.SubjectDetails {
	return &gencomment.SubjectDetails{
		Id:           data.ID.Hex(),
		ItemId:       data.ItemId,
		UserId:       data.UserId,
		TopCommentId: *data.TopCommentId,
		RootCount:    *data.RootCount,
		AllCount:     *data.AllCount,
		State:        data.State,
		Attrs:        data.Attrs,
	}
}

func ParsePagination(opts *basic.PaginationOptions) (p *pagination.PaginationOptions) {
	if opts == nil {
		p = &pagination.PaginationOptions{}
	} else {
		p = &pagination.PaginationOptions{
			Limit:     opts.Limit,
			Offset:    opts.Offset,
			Backward:  opts.Backward,
			LastToken: opts.LastToken,
		}
	}
	return
}

func LabelMapperToLabel(data *label.Label) *gencomment.Label {
	return &gencomment.Label{
		LabelId: data.ID.Hex(),
		Value:   data.Value,
	}
}

func LabelToLabelMapper(data *gencomment.Label) *label.Label {
	oid, _ := primitive.ObjectIDFromHex(data.LabelId)
	return &label.Label{
		ID:    oid,
		Value: data.Value,
	}
}

func LabelEntityToLabelEntityMapper(data *gencomment.LabelEntity) *labelEntity.LabelEntity {
	oid, _ := primitive.ObjectIDFromHex(data.ObjectId)
	return &labelEntity.LabelEntity{
		ID:         oid,
		ObjectType: data.ObjectType,
		Labels:     data.Labels,
	}
}

func LabelEntityFilterOptionsToFilterOptions(data *gencomment.ObjectFilterOptions) *labelEntity.FilterOptions {
	if data == nil {
		return &labelEntity.FilterOptions{}
	} else {
		return &labelEntity.FilterOptions{
			OnlyLabelId:    data.OnlyLabelId,
			OnlyObjectType: data.OnlyObjectType,
		}
	}
}

func ConvertLabelAllFieldsSearchQuery(data string) []types.Query {
	return []types.Query{{
		MultiMatch: &types.MultiMatchQuery{
			Query:  data,
			Fields: []string{consts.Value + "^3", consts.ID},
		}},
	}
}
