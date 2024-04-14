package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform/biz/infrastructure/convertor"
	commentMapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/comment"
	subjectMapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/subject"
	"github.com/CloudStriver/platform/biz/infrastructure/sort"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/google/wire"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
)

type ICommentService interface {
	UpdateCount(ctx context.Context, rootId string, count int64)
	GetComment(ctx context.Context, req *platform.GetCommentReq) (resp *platform.GetCommentResp, err error)
	GetCommentList(ctx context.Context, req *platform.GetCommentListReq) (resp *platform.GetCommentListResp, err error)
	GetCommentBlocks(ctx context.Context, req *platform.GetCommentBlocksReq) (resp *platform.GetCommentBlocksResp, err error)
	CreateComment(ctx context.Context, req *platform.CreateCommentReq) (resp *platform.CreateCommentResp, err error)
	UpdateComment(ctx context.Context, req *platform.UpdateCommentReq) (resp *platform.UpdateCommentResp, err error)
	DeleteComment(ctx context.Context, commentId string, level bool) (resp *platform.DeleteCommentResp, err error)
	DeleteCommentByIds(ctx context.Context, req *platform.DeleteCommentByIdsReq) (resp *platform.DeleteCommentByIdsResp, err error)
	SetCommentAttrs(ctx context.Context, req *platform.SetCommentAttrsReq, res *platform.GetCommentSubjectResp) (resp *platform.SetCommentAttrsResp, err error)
}

type CommentService struct {
	CommentMongoMapper commentMapper.IMongoMapper
	SubjectMongoMapper subjectMapper.IMongoMapper
}

var CommentSet = wire.NewSet(
	wire.Struct(new(CommentService), "*"),
	wire.Bind(new(ICommentService), new(*CommentService)),
)

func (s *CommentService) DeleteCommentByIds(ctx context.Context, req *platform.DeleteCommentByIdsReq) (resp *platform.DeleteCommentByIdsResp, err error) {
	resp = new(platform.DeleteCommentByIdsResp)
	if _, err = s.CommentMongoMapper.DeleteMany(ctx, req.CommentIds); err != nil {
		log.CtxError(ctx, "删除评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) GetComment(ctx context.Context, req *platform.GetCommentReq) (resp *platform.GetCommentResp, err error) {
	resp = new(platform.GetCommentResp)
	var data *commentMapper.Comment
	if data, err = s.CommentMongoMapper.FindOne(ctx, req.CommentId); err != nil {
		log.CtxError(ctx, "获取评论详情 失败[%v]\n", err)
		return resp, err
	}

	resp = &platform.GetCommentResp{
		SubjectId:  data.SubjectId,
		RootId:     data.RootId,
		FatherId:   data.FatherId,
		Count:      *data.Count,
		State:      data.State,
		Attrs:      data.Attrs,
		LabelIds:   data.Labels,
		UserId:     data.UserId,
		AtUserId:   data.AtUserId,
		Content:    data.Content,
		Meta:       data.Meta,
		CreateTime: data.CreateAt.UnixMilli(),
	}
	return resp, nil
}

func (s *CommentService) GetCommentList(ctx context.Context, req *platform.GetCommentListReq) (resp *platform.GetCommentListResp, err error) {
	resp = new(platform.GetCommentListResp)
	var (
		total    int64
		comments []*commentMapper.Comment
	)

	p := convertor.ParsePagination(req.Pagination)
	filter := convertor.CommentFilterOptionsToFilterOptions(req.FilterOptions)
	if comments, total, err = s.CommentMongoMapper.FindManyAndCount(ctx, filter, p, sort.TimeCursorType); err != nil {
		log.CtxError(ctx, "获取评论列表 失败[%v]\n", err)
		return resp, err
	}
	if p.LastToken != nil {
		resp.Token = *p.LastToken
	}
	resp.Comments = lo.Map(comments, func(comment *commentMapper.Comment, _ int) *platform.Comment {
		return convertor.CommentMapperToComment(comment)
	})
	resp.Total = total
	return resp, nil
}

func (s *CommentService) GetCommentBlocks(ctx context.Context, req *platform.GetCommentBlocksReq) (resp *platform.GetCommentBlocksResp, err error) {
	resp = new(platform.GetCommentBlocksResp)

	var (
		total     int64
		comments  []*commentMapper.Comment
		replyList []*commentMapper.Comment
		filter    *commentMapper.FilterOptions
	)

	p := convertor.ParsePagination(req.Pagination)
	filter = &commentMapper.FilterOptions{OnlyRootId: lo.ToPtr(req.RootId)}
	if req.RootId == req.SubjectId {
		if comments, total, err = s.CommentMongoMapper.FindManyAndCount(ctx, filter, p, sort.TimeCursorType); err != nil {
			log.CtxError(ctx, "获取评论列表 失败[%v]\n", err)
			return resp, err
		}
		if p.LastToken != nil {
			resp.Token = *p.LastToken
		}
		resp.Total = total
		resp.CommentBlocks = lo.Map(comments, func(comment *commentMapper.Comment, _ int) *platform.CommentBlock {
			return &platform.CommentBlock{
				RootComment: convertor.CommentMapperToComment(comment),
				ReplyList:   &platform.ReplyList{},
			}
		})

		for i, comment := range comments {
			p = &pagination.PaginationOptions{}
			filter = &commentMapper.FilterOptions{OnlyRootId: lo.ToPtr(comment.ID.Hex())}
			if replyList, total, err = s.CommentMongoMapper.FindManyAndCount(ctx, filter, p, sort.TimeCursorType); err != nil {
				log.CtxError(ctx, "获取评论列表 失败[%v]\n", err)
				return resp, err
			}

			if p.LastToken != nil {
				resp.CommentBlocks[i].ReplyList.Token = *p.LastToken
			}
			resp.CommentBlocks[i].ReplyList.Total = total
			resp.CommentBlocks[i].ReplyList.Comments = lo.Map(replyList, func(comment *commentMapper.Comment, _ int) *platform.Comment {
				return convertor.CommentMapperToComment(comment)
			})
		}
	} else {
		if comments, total, err = s.CommentMongoMapper.FindManyAndCount(ctx, filter, p, sort.TimeCursorType); err != nil {
			log.CtxError(ctx, "获取评论列表 失败[%v]\n", err)
			return resp, err
		}
		resp.CommentBlocks = make([]*platform.CommentBlock, 1)
		resp.CommentBlocks[0] = &platform.CommentBlock{
			ReplyList: &platform.ReplyList{},
		}
		if p.LastToken != nil {
			resp.CommentBlocks[0].ReplyList.Token = *p.LastToken
		}
		resp.CommentBlocks[0].ReplyList.Total = total
		resp.CommentBlocks[0].ReplyList.Comments = lo.Map(comments, func(comment *commentMapper.Comment, _ int) *platform.Comment {
			return convertor.CommentMapperToComment(comment)
		})
	}
	return resp, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *platform.CreateCommentReq) (resp *platform.CreateCommentResp, err error) {
	resp = new(platform.CreateCommentResp)
	if resp.CommentId, err = s.CommentMongoMapper.Insert(ctx, &commentMapper.Comment{
		ID:        primitive.NilObjectID,
		UserId:    req.UserId,
		AtUserId:  req.AtUserId,
		SubjectId: req.SubjectId,
		RootId:    req.RootId,
		FatherId:  req.FatherId,
		Content:   req.Content,
		Meta:      req.Meta,
		Labels:    req.LabelIds,
		Count:     lo.ToPtr(int64(0)),
		State:     int64(platform.State_Normal),
		Attrs:     int64(platform.Attrs_None),
	}); err != nil {
		log.CtxError(ctx, "创建评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) UpdateCount(ctx context.Context, rootId string, count int64) {
	s.CommentMongoMapper.UpdateCount(ctx, rootId, count)
}

func (s *CommentService) UpdateComment(ctx context.Context, req *platform.UpdateCommentReq) (resp *platform.UpdateCommentResp, err error) {
	resp = new(platform.UpdateCommentResp)
	var oid primitive.ObjectID
	if oid, err = primitive.ObjectIDFromHex(req.CommentId); err != nil {
		return resp, err
	}
	if _, err = s.CommentMongoMapper.Update(ctx, &commentMapper.Comment{
		ID:     oid,
		Meta:   req.Meta,
		Labels: req.LabelIds,
		State:  req.State,
	}); err != nil {
		log.CtxError(ctx, "更新评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentId string, level bool) (resp *platform.DeleteCommentResp, err error) {
	resp = new(platform.DeleteCommentResp)
	if level {

	} else {
		if _, err = s.CommentMongoMapper.Delete(ctx, commentId); err != nil {
			log.CtxError(ctx, "删除评论 失败[%v]\n", err)
			return resp, err
		}
	}
	return resp, nil
}

func (s *CommentService) SetCommentAttrs(ctx context.Context, req *platform.SetCommentAttrsReq, res *platform.GetCommentSubjectResp) (resp *platform.SetCommentAttrsResp, err error) {
	resp = new(platform.SetCommentAttrsResp)
	var (
		subjectId primitive.ObjectID
		commentId primitive.ObjectID
	)

	if subjectId, err = primitive.ObjectIDFromHex(req.SubjectId); err != nil {
		return resp, err
	}
	if commentId, err = primitive.ObjectIDFromHex(req.CommentId); err != nil {
		return resp, err
	}

	switch req.Attrs {
	case int64(platform.Attrs_Pinned):
		req.SortTime = math.MaxInt64 - 1
	case int64(platform.Attrs_PinnedAndHighlighted):
		req.SortTime = math.MaxInt64 - 1
	}

	data := &commentMapper.Comment{ID: commentId, Attrs: req.Attrs, SortTime: req.SortTime}
	tx := s.SubjectMongoMapper.StartClient()
	if res.TopCommentId == "" {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == int64(platform.Attrs_Pinned) || req.Attrs == int64(platform.Attrs_PinnedAndHighlighted) {
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: subjectId, TopCommentId: lo.ToPtr(req.CommentId)}); err != nil {
					if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
						log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
						return err
					}
				}
			}
			if _, err = s.CommentMongoMapper.Update(sessionContext, data); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
					return err
				}
			}
			if err = sessionContext.CommitTransaction(sessionContext); err != nil {
				log.CtxError(sessionContext, "设置评论属性: 提交事务异常[%v]\n", err)
				return err
			}
			return nil
		})
	} else if res.TopCommentId == req.CommentId {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == int64(platform.Attrs_None) || req.Attrs == int64(platform.Attrs_Highlighted) {
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: subjectId, TopCommentId: lo.ToPtr("")}); err != nil {
					if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
						log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
						return err
					}
				}
			}
			if _, err = s.CommentMongoMapper.Update(sessionContext, data); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
					return err
				}
			}
			if err = sessionContext.CommitTransaction(sessionContext); err != nil {
				log.CtxError(sessionContext, "设置评论属性: 提交事务异常[%v]\n", err)
				return err
			}
			return nil
		})
	} else {
		var oldComment *commentMapper.Comment
		if oldComment, err = s.CommentMongoMapper.FindOne(ctx, res.TopCommentId); err != nil {
			return resp, err
		}
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			var subject *subjectMapper.Subject
			switch {
			case req.Attrs == int64(platform.Attrs_Pinned) || req.Attrs == int64(platform.Attrs_PinnedAndHighlighted):
				subject = &subjectMapper.Subject{ID: subjectId, TopCommentId: lo.ToPtr(req.CommentId)}
			default:
				subject = &subjectMapper.Subject{ID: subjectId, TopCommentId: lo.ToPtr("")}
			}
			if _, err = s.SubjectMongoMapper.Update(sessionContext, subject); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)

					return err
				}
			}
			oid, _ := primitive.ObjectIDFromHex(res.TopCommentId)
			oldData := &commentMapper.Comment{ID: oid, Attrs: int64(platform.Attrs_None), SortTime: oldComment.CreateAt.UnixMilli()}
			if _, err = s.CommentMongoMapper.Update(sessionContext, oldData); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
					return err
				}
			}
			if _, err = s.CommentMongoMapper.Update(sessionContext, data); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
				}
				return err
			}
			if err = sessionContext.CommitTransaction(sessionContext); err != nil {
				log.CtxError(sessionContext, "设置评论属性: 提交事务异常[%v]\n", err)
				return err
			}
			return nil
		})
	}
	return resp, err
}
