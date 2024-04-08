package convertor

import (
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/CloudStriver/platform/biz/infrastructure/mapper/comment"
	"github.com/CloudStriver/platform/biz/infrastructure/mapper/label"
	"github.com/CloudStriver/platform/biz/infrastructure/mapper/subject"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/basic"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func CommentMapperToCommentInfo(data *comment.Comment) *platform.Comment {
	return &platform.Comment{
		Id:         data.ID.Hex(),
		SubjectId:  data.SubjectId,
		RootId:     data.RootId,
		FatherId:   data.FatherId,
		Count:      *data.Count,
		State:      data.State,
		Attrs:      data.Attrs,
		Labels:     data.Labels,
		UserId:     data.UserId,
		AtUserId:   data.AtUserId,
		Content:    data.Content,
		Meta:       data.Meta,
		CreateTime: data.CreateAt.UnixMilli(),
	}
}

func CommentFilterOptionsToFilterOptions(data *platform.CommentFilterOptions) *comment.FilterOptions {
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

func SubjectMapperToSubjectDetail(data *subject.Subject) *platform.Subject {
	return &platform.Subject{
		Id:           data.ID.Hex(),
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

func LabelMapperToLabel(data *label.Label) *platform.Label {
	return &platform.Label{
		Id:       data.ID.Hex(),
		FatherId: data.FatherId,
		Value:    data.Value,
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
