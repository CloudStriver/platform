package adaptor

import (
	"context"
	"github.com/CloudStriver/platform-comment/biz/application/service"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	genrelation "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/relation"
	"github.com/zeromicro/go-zero/core/mr"
)

type PlatformServerImpl struct {
	*config.Config
	CommentService  service.ICommentService
	LabelService    service.ILabelService
	SubjectService  service.ISubjectService
	RelationService service.RelationService
}

func (s *PlatformServerImpl) DeleteNode(ctx context.Context, req *genrelation.DeleteNodeReq) (res *genrelation.DeleteNodeResp, err error) {
	return s.RelationService.DeleteNode(ctx, req)
}

func (s *PlatformServerImpl) GetRelationPaths(ctx context.Context, req *genrelation.GetRelationPathsReq) (res *genrelation.GetRelationPathsResp, err error) {
	return s.RelationService.GetRelationPaths(ctx, req)
}

func (s *PlatformServerImpl) GetRelationCount(ctx context.Context, req *genrelation.GetRelationCountReq) (res *genrelation.GetRelationCountResp, err error) {
	return s.RelationService.GetRelationCount(ctx, req)
}

func (s *PlatformServerImpl) GetRelations(ctx context.Context, req *genrelation.GetRelationsReq) (resp *genrelation.GetRelationsResp, err error) {
	return s.RelationService.GetRelations(ctx, req)
}

func (s *PlatformServerImpl) DeleteRelation(ctx context.Context, req *genrelation.DeleteRelationReq) (resp *genrelation.DeleteRelationResp, err error) {
	return s.RelationService.DeleteRelation(ctx, req)
}

func (s *PlatformServerImpl) CreateRelation(ctx context.Context, req *genrelation.CreateRelationReq) (resp *genrelation.CreateRelationResp, err error) {
	return s.RelationService.CreateRelation(ctx, req)
}

func (s *PlatformServerImpl) GetRelation(ctx context.Context, req *genrelation.GetRelationReq) (resp *genrelation.GetRelationResp, err error) {
	return s.RelationService.GetRelation(ctx, req)
}

func (c *PlatformServerImpl) DeleteCommentByIds(ctx context.Context, req *comment.DeleteCommentByIdsReq) (res *comment.DeleteCommentByIdsResp, err error) {
	return c.CommentService.DeleteCommentByIds(ctx, req)
}

func (c *PlatformServerImpl) GetComment(ctx context.Context, req *comment.GetCommentReq) (res *comment.GetCommentResp, err error) {
	return c.CommentService.GetComment(ctx, req)
}

func (c *PlatformServerImpl) GetCommentList(ctx context.Context, req *comment.GetCommentListReq) (res *comment.GetCommentListResp, err error) {
	return c.CommentService.GetCommentList(ctx, req)
}

func (c *PlatformServerImpl) CreateComment(ctx context.Context, req *comment.CreateCommentReq) (res *comment.CreateCommentResp, err error) {
	if res, err = c.CommentService.CreateComment(ctx, req); err != nil {
		return res, err
	}
	_ = mr.Finish(func() error {
		c.CommentService.UpdateCount(ctx, req.Comment.RootId, req.Comment.SubjectId, req.Comment.FatherId, consts.Increment)
		return nil
	}, func() error {
		c.SubjectService.UpdateCount(ctx, req.Comment.RootId, req.Comment.SubjectId, req.Comment.FatherId, consts.Increment)
		return nil
	})
	return res, nil
}

func (c *PlatformServerImpl) UpdateComment(ctx context.Context, req *comment.UpdateCommentReq) (res *comment.UpdateCommentResp, err error) {
	return c.CommentService.UpdateComment(ctx, req)
}

func (c *PlatformServerImpl) DeleteComment(ctx context.Context, req *comment.DeleteCommentReq) (res *comment.DeleteCommentResp, err error) {
	var data *comment.GetCommentResp
	if data, err = c.CommentService.GetComment(ctx, &comment.GetCommentReq{CommentId: req.Id}); err != nil {
		return res, err
	}
	if res, err = c.CommentService.DeleteComment(ctx, req); err != nil {
		return res, err
	}
	_ = mr.Finish(func() error {
		c.CommentService.UpdateCount(ctx, data.Comment.RootId, data.Comment.SubjectId, data.Comment.FatherId, consts.Decrement)
		return nil
	}, func() error {
		c.SubjectService.UpdateCount(ctx, data.Comment.RootId, data.Comment.SubjectId, data.Comment.FatherId, consts.Decrement)
		return nil
	})
	return res, nil
}

func (c *PlatformServerImpl) SetCommentAttrs(ctx context.Context, req *comment.SetCommentAttrsReq) (res *comment.SetCommentAttrsResp, err error) {
	var resp *comment.GetCommentSubjectResp
	if resp, err = c.SubjectService.GetCommentSubject(ctx, &comment.GetCommentSubjectReq{Id: req.SubjectId}); err != nil {
		return res, err
	}
	return c.CommentService.SetCommentAttrs(ctx, req, resp)
}

func (c *PlatformServerImpl) GetCommentSubject(ctx context.Context, req *comment.GetCommentSubjectReq) (res *comment.GetCommentSubjectResp, err error) {
	return c.SubjectService.GetCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) CreateCommentSubject(ctx context.Context, req *comment.CreateCommentSubjectReq) (res *comment.CreateCommentSubjectResp, err error) {
	return c.SubjectService.CreateCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) UpdateCommentSubject(ctx context.Context, req *comment.UpdateCommentSubjectReq) (res *comment.UpdateCommentSubjectResp, err error) {
	return c.SubjectService.UpdateCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) DeleteCommentSubject(ctx context.Context, req *comment.DeleteCommentSubjectReq) (res *comment.DeleteCommentSubjectResp, err error) {
	return c.SubjectService.DeleteCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) CreateLabel(ctx context.Context, req *comment.CreateLabelReq) (res *comment.CreateLabelResp, err error) {
	return c.LabelService.CreateLabel(ctx, req)
}

func (c *PlatformServerImpl) DeleteLabel(ctx context.Context, req *comment.DeleteLabelReq) (res *comment.DeleteLabelResp, err error) {
	return c.LabelService.DeleteLabel(ctx, req)
}

func (c *PlatformServerImpl) GetLabel(ctx context.Context, req *comment.GetLabelReq) (res *comment.GetLabelResp, err error) {
	return c.LabelService.GetLabel(ctx, req)
}

func (c *PlatformServerImpl) GetLabels(ctx context.Context, req *comment.GetLabelsReq) (res *comment.GetLabelsResp, err error) {
	return c.LabelService.GetLabels(ctx, req)
}

func (c *PlatformServerImpl) GetLabelsInBatch(ctx context.Context, req *comment.GetLabelsInBatchReq) (res *comment.GetLabelsInBatchResp, err error) {
	return c.LabelService.GetLabelsInBatch(ctx, req)
}

func (c *PlatformServerImpl) UpdateLabel(ctx context.Context, req *comment.UpdateLabelReq) (res *comment.UpdateLabelResp, err error) {
	return c.LabelService.UpdateLabel(ctx, req)
}
