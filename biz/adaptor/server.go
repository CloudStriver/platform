package adaptor

import (
	"context"
	"github.com/CloudStriver/platform/biz/application/service"
	"github.com/CloudStriver/platform/biz/infrastructure/config"
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/zeromicro/go-zero/core/mr"
)

type PlatformServerImpl struct {
	*config.Config
	CommentService  service.ICommentService
	LabelService    service.ILabelService
	SubjectService  service.ISubjectService
	RelationService service.RelationService
}

func (s *PlatformServerImpl) GetRelationPathsCount(ctx context.Context, req *platform.GetRelationPathsCountReq) (res *platform.GetRelationPathsCountResp, err error) {
	return s.RelationService.GetRelationPathsCount(ctx, req)
}

func (s *PlatformServerImpl) DeleteNode(ctx context.Context, req *platform.DeleteNodeReq) (res *platform.DeleteNodeResp, err error) {
	return s.RelationService.DeleteNode(ctx, req)
}

func (s *PlatformServerImpl) GetRelationPaths(ctx context.Context, req *platform.GetRelationPathsReq) (res *platform.GetRelationPathsResp, err error) {
	return s.RelationService.GetRelationPaths(ctx, req)
}

func (s *PlatformServerImpl) GetRelationCount(ctx context.Context, req *platform.GetRelationCountReq) (res *platform.GetRelationCountResp, err error) {
	return s.RelationService.GetRelationCount(ctx, req)
}

func (s *PlatformServerImpl) GetRelations(ctx context.Context, req *platform.GetRelationsReq) (resp *platform.GetRelationsResp, err error) {
	return s.RelationService.GetRelations(ctx, req)
}

func (s *PlatformServerImpl) DeleteRelation(ctx context.Context, req *platform.DeleteRelationReq) (resp *platform.DeleteRelationResp, err error) {
	return s.RelationService.DeleteRelation(ctx, req)
}

func (s *PlatformServerImpl) CreateRelation(ctx context.Context, req *platform.CreateRelationReq) (resp *platform.CreateRelationResp, err error) {
	return s.RelationService.CreateRelation(ctx, req)
}

func (s *PlatformServerImpl) GetRelation(ctx context.Context, req *platform.GetRelationReq) (resp *platform.GetRelationResp, err error) {
	return s.RelationService.GetRelation(ctx, req)
}

func (c *PlatformServerImpl) DeleteCommentByIds(ctx context.Context, req *platform.DeleteCommentByIdsReq) (res *platform.DeleteCommentByIdsResp, err error) {
	return c.CommentService.DeleteCommentByIds(ctx, req)
}

func (c *PlatformServerImpl) GetComment(ctx context.Context, req *platform.GetCommentReq) (res *platform.GetCommentResp, err error) {
	return c.CommentService.GetComment(ctx, req)
}

func (c *PlatformServerImpl) GetCommentList(ctx context.Context, req *platform.GetCommentListReq) (res *platform.GetCommentListResp, err error) {
	return c.CommentService.GetCommentList(ctx, req)
}

func (c *PlatformServerImpl) CreateComment(ctx context.Context, req *platform.CreateCommentReq) (res *platform.CreateCommentResp, err error) {
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

func (c *PlatformServerImpl) UpdateComment(ctx context.Context, req *platform.UpdateCommentReq) (res *platform.UpdateCommentResp, err error) {
	return c.CommentService.UpdateComment(ctx, req)
}

func (c *PlatformServerImpl) DeleteComment(ctx context.Context, req *platform.DeleteCommentReq) (res *platform.DeleteCommentResp, err error) {
	var data *platform.GetCommentResp
	if data, err = c.CommentService.GetComment(ctx, &platform.GetCommentReq{CommentId: req.Id}); err != nil {
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

func (c *PlatformServerImpl) SetCommentAttrs(ctx context.Context, req *platform.SetCommentAttrsReq) (res *platform.SetCommentAttrsResp, err error) {
	var resp *platform.GetCommentSubjectResp
	if resp, err = c.SubjectService.GetCommentSubject(ctx, &platform.GetCommentSubjectReq{Id: req.SubjectId}); err != nil {
		return res, err
	}
	return c.CommentService.SetCommentAttrs(ctx, req, resp)
}

func (c *PlatformServerImpl) GetCommentSubject(ctx context.Context, req *platform.GetCommentSubjectReq) (res *platform.GetCommentSubjectResp, err error) {
	return c.SubjectService.GetCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) CreateCommentSubject(ctx context.Context, req *platform.CreateCommentSubjectReq) (res *platform.CreateCommentSubjectResp, err error) {
	return c.SubjectService.CreateCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) UpdateCommentSubject(ctx context.Context, req *platform.UpdateCommentSubjectReq) (res *platform.UpdateCommentSubjectResp, err error) {
	return c.SubjectService.UpdateCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) DeleteCommentSubject(ctx context.Context, req *platform.DeleteCommentSubjectReq) (res *platform.DeleteCommentSubjectResp, err error) {
	return c.SubjectService.DeleteCommentSubject(ctx, req)
}

func (c *PlatformServerImpl) CreateLabel(ctx context.Context, req *platform.CreateLabelReq) (res *platform.CreateLabelResp, err error) {
	return c.LabelService.CreateLabel(ctx, req)
}

func (c *PlatformServerImpl) DeleteLabel(ctx context.Context, req *platform.DeleteLabelReq) (res *platform.DeleteLabelResp, err error) {
	return c.LabelService.DeleteLabel(ctx, req)
}

func (c *PlatformServerImpl) GetLabel(ctx context.Context, req *platform.GetLabelReq) (res *platform.GetLabelResp, err error) {
	return c.LabelService.GetLabel(ctx, req)
}

func (c *PlatformServerImpl) GetLabels(ctx context.Context, req *platform.GetLabelsReq) (res *platform.GetLabelsResp, err error) {
	return c.LabelService.GetLabels(ctx, req)
}

func (c *PlatformServerImpl) GetLabelsInBatch(ctx context.Context, req *platform.GetLabelsInBatchReq) (res *platform.GetLabelsInBatchResp, err error) {
	return c.LabelService.GetLabelsInBatch(ctx, req)
}

func (c *PlatformServerImpl) UpdateLabel(ctx context.Context, req *platform.UpdateLabelReq) (res *platform.UpdateLabelResp, err error) {
	return c.LabelService.UpdateLabel(ctx, req)
}
