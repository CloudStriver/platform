package adaptor

import (
	"context"
	"github.com/CloudStriver/platform-comment/biz/application/service"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	"github.com/zeromicro/go-zero/core/mr"
)

type CommentServerImpl struct {
	*config.Config
	CommentService service.ICommentService
	LabelService   service.ILabelService
	SubjectService service.ISubjectService
}

func (c *CommentServerImpl) GetComment(ctx context.Context, req *comment.GetCommentReq) (res *comment.GetCommentResp, err error) {
	return c.CommentService.GetComment(ctx, req)
}

func (c *CommentServerImpl) GetCommentList(ctx context.Context, req *comment.GetCommentListReq) (res *comment.GetCommentListResp, err error) {
	return c.CommentService.GetCommentList(ctx, req)
}

func (c *CommentServerImpl) CreateComment(ctx context.Context, req *comment.CreateCommentReq) (res *comment.CreateCommentResp, err error) {
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

func (c *CommentServerImpl) UpdateComment(ctx context.Context, req *comment.UpdateCommentReq) (res *comment.UpdateCommentResp, err error) {
	return c.CommentService.UpdateComment(ctx, req)
}

func (c *CommentServerImpl) DeleteComment(ctx context.Context, req *comment.DeleteCommentReq) (res *comment.DeleteCommentResp, err error) {
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

func (c *CommentServerImpl) SetCommentAttrs(ctx context.Context, req *comment.SetCommentAttrsReq) (res *comment.SetCommentAttrsResp, err error) {
	var resp *comment.GetCommentSubjectResp
	if resp, err = c.SubjectService.GetCommentSubject(ctx, &comment.GetCommentSubjectReq{Id: req.SubjectId}); err != nil {
		return res, err
	}
	return c.CommentService.SetCommentAttrs(ctx, req, resp)
}

func (c *CommentServerImpl) GetCommentSubject(ctx context.Context, req *comment.GetCommentSubjectReq) (res *comment.GetCommentSubjectResp, err error) {
	return c.SubjectService.GetCommentSubject(ctx, req)
}

func (c *CommentServerImpl) CreateCommentSubject(ctx context.Context, req *comment.CreateCommentSubjectReq) (res *comment.CreateCommentSubjectResp, err error) {
	return c.SubjectService.CreateCommentSubject(ctx, req)
}

func (c *CommentServerImpl) UpdateCommentSubject(ctx context.Context, req *comment.UpdateCommentSubjectReq) (res *comment.UpdateCommentSubjectResp, err error) {
	return c.SubjectService.UpdateCommentSubject(ctx, req)
}

func (c *CommentServerImpl) DeleteCommentSubject(ctx context.Context, req *comment.DeleteCommentSubjectReq) (res *comment.DeleteCommentSubjectResp, err error) {
	return c.SubjectService.DeleteCommentSubject(ctx, req)
}

func (c *CommentServerImpl) CreateLabel(ctx context.Context, req *comment.CreateLabelReq) (res *comment.CreateLabelResp, err error) {
	return c.LabelService.CreateLabel(ctx, req)
}

func (c *CommentServerImpl) DeleteLabel(ctx context.Context, req *comment.DeleteLabelReq) (res *comment.DeleteLabelResp, err error) {
	return c.LabelService.DeleteLabel(ctx, req)
}

func (c *CommentServerImpl) GetLabel(ctx context.Context, req *comment.GetLabelReq) (res *comment.GetLabelResp, err error) {
	return c.LabelService.GetLabel(ctx, req)
}

func (c *CommentServerImpl) GetLabels(ctx context.Context, req *comment.GetLabelsReq) (res *comment.GetLabelsResp, err error) {
	return c.LabelService.GetLabels(ctx, req)
}

func (c *CommentServerImpl) GetLabelsInBatch(ctx context.Context, req *comment.GetLabelsInBatchReq) (res *comment.GetLabelsInBatchResp, err error) {
	return c.LabelService.GetLabelsInBatch(ctx, req)
}

func (c *CommentServerImpl) UpdateLabel(ctx context.Context, req *comment.UpdateLabelReq) (res *comment.UpdateLabelResp, err error) {
	return c.LabelService.UpdateLabel(ctx, req)
}
