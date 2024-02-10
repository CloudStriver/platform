package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/convertor"
	commentMapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/comment"
	subjectMapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/subject"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/sort"
	gencomment "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	"github.com/google/wire"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
)

type ICommentService interface {
	UpdateAfterCreateComment(ctx context.Context, req *gencomment.Comment)
	GetComment(ctx context.Context, req *gencomment.GetCommentReq) (resp *gencomment.GetCommentResp, err error)
	GetCommentList(ctx context.Context, req *gencomment.GetCommentListReq) (resp *gencomment.GetCommentListResp, err error)
	CreateComment(ctx context.Context, req *gencomment.CreateCommentReq) (resp *gencomment.CreateCommentResp, err error)
	UpdateComment(ctx context.Context, req *gencomment.UpdateCommentReq) (resp *gencomment.UpdateCommentResp, err error)
	DeleteComment(ctx context.Context, req *gencomment.DeleteCommentReq) (resp *gencomment.DeleteCommentResp, err error)
	DeleteCommentWithUserId(ctx context.Context, req *gencomment.DeleteCommentWithUserIdReq) (resp *gencomment.DeleteCommentWithUserIdResp, err error)
	SetCommentState(ctx context.Context, req *gencomment.SetCommentStateReq) (resp *gencomment.SetCommentStateResp, err error)
	SetCommentAttrs(ctx context.Context, req *gencomment.SetCommentAttrsReq, res *gencomment.GetCommentSubjectResp) (resp *gencomment.SetCommentAttrsResp, err error)
}

type CommentService struct {
	CommentMongoMapper commentMapper.IMongoMapper
	SubjectMongoMapper subjectMapper.IMongoMapper
}

var CommentSet = wire.NewSet(
	wire.Struct(new(CommentService), "*"),
	wire.Bind(new(ICommentService), new(*CommentService)),
)

func (s *CommentService) GetComment(ctx context.Context, req *gencomment.GetCommentReq) (resp *gencomment.GetCommentResp, err error) {
	resp = new(gencomment.GetCommentResp)
	var data *commentMapper.Comment
	if data, err = s.CommentMongoMapper.FindOne(ctx, req.CommentId); err != nil {
		log.CtxError(ctx, "获取评论详情 失败[%v]\n", err)
		return resp, err
	}
	resp.Comment = convertor.CommentMapperToCommentInfo(data)
	return resp, nil
}

func (s *CommentService) GetCommentList(ctx context.Context, req *gencomment.GetCommentListReq) (resp *gencomment.GetCommentListResp, err error) {
	resp = new(gencomment.GetCommentListResp)
	var total int64
	var comments []*commentMapper.Comment

	p := convertor.ParsePagination(req.Pagination)
	filter := convertor.CommentFilterOptionsToFilterOptions(req.FilterOptions)
	if comments, total, err = s.CommentMongoMapper.FindManyAndCount(ctx, filter, p, sort.TimeCursorType); err != nil {
		log.CtxError(ctx, "获取评论列表 失败[%v]\n", err)
		return resp, err
	}
	if p.LastToken != nil {
		resp.Token = *p.LastToken
	}
	resp.Comments = lo.Map(comments, func(comment *commentMapper.Comment, _ int) *gencomment.CommentInfo {
		return convertor.CommentMapperToCommentInfo(comment)
	})
	resp.Total = total
	return resp, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *gencomment.CreateCommentReq) (resp *gencomment.CreateCommentResp, err error) {
	resp = new(gencomment.CreateCommentResp)
	data := convertor.CommentToCommentMapper(req.Comment)
	if resp.Id, err = s.CommentMongoMapper.Insert(ctx, data); err != nil {
		log.CtxError(ctx, "创建评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) UpdateAfterCreateComment(ctx context.Context, req *gencomment.Comment) {
	if req.RootId != req.SubjectId {
		if req.FatherId != req.SubjectId {
			// 二级评论 + 三级评论
			data := convertor.CommentToCommentMapper(&gencomment.Comment{Id: req.RootId, Count: lo.ToPtr(int64(1))})
			s.CommentMongoMapper.UpdateAfterCreateComment(ctx, data)
		}
	}
}

func (s *CommentService) UpdateComment(ctx context.Context, req *gencomment.UpdateCommentReq) (resp *gencomment.UpdateCommentResp, err error) {
	resp = new(gencomment.UpdateCommentResp)
	data := convertor.CommentToCommentMapper(req.Comment)
	if _, err = s.CommentMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "更新评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *gencomment.DeleteCommentReq) (resp *gencomment.DeleteCommentResp, err error) {
	resp = new(gencomment.DeleteCommentResp)
	if _, err = s.CommentMongoMapper.Delete(ctx, req.Id); err != nil {
		log.CtxError(ctx, "删除评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) DeleteCommentWithUserId(ctx context.Context, req *gencomment.DeleteCommentWithUserIdReq) (resp *gencomment.DeleteCommentWithUserIdResp, err error) {
	resp = new(gencomment.DeleteCommentWithUserIdResp)
	if _, err = s.CommentMongoMapper.DeleteWithUserId(ctx, req.Id, req.UserId); err != nil {
		log.CtxError(ctx, "删除评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}
func (s *CommentService) SetCommentState(ctx context.Context, req *gencomment.SetCommentStateReq) (resp *gencomment.SetCommentStateResp, err error) {
	resp = new(gencomment.SetCommentStateResp)
	data := convertor.CommentToCommentMapper(&gencomment.Comment{Id: req.Id, UserId: req.UserId, State: req.State})
	if _, err = s.CommentMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "设置评论状态 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) SetCommentAttrs(ctx context.Context, req *gencomment.SetCommentAttrsReq, res *gencomment.GetCommentSubjectResp) (resp *gencomment.SetCommentAttrsResp, err error) {
	resp = new(gencomment.SetCommentAttrsResp)
	if req.Attrs == consts.Pinned || req.Attrs == consts.PinnedAndHighlighted {
		req.SortTime = math.MaxInt64 - 1
	}
	tx := s.SubjectMongoMapper.StartClient()
	if res.Subject.TopCommentId == "" {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == consts.Pinned || req.Attrs == consts.PinnedAndHighlighted {
				oid, _ := primitive.ObjectIDFromHex(req.SubjectId)
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: oid, UserId: req.UserId, TopCommentId: lo.ToPtr(req.Id)}); err != nil {
					if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
						log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
						return err
					}
				}
			}
			data := convertor.CommentToCommentMapper(&gencomment.Comment{Id: req.Id, UserId: req.UserId, Attrs: req.Attrs, CreateTime: req.SortTime})
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
	} else if res.Subject.TopCommentId == req.Id {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == consts.None {
				oid, _ := primitive.ObjectIDFromHex(req.SubjectId)
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: oid, UserId: req.UserId, TopCommentId: lo.ToPtr("")}); err != nil {
					if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
						log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
						return err
					}
				}
			}

			data := convertor.CommentToCommentMapper(&gencomment.Comment{Id: req.Id, UserId: req.UserId, Attrs: req.Attrs, CreateTime: req.SortTime})
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
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == consts.Pinned || req.Attrs == consts.PinnedAndHighlighted || req.Attrs == consts.None {
				oid, _ := primitive.ObjectIDFromHex(req.SubjectId)
				var data *subjectMapper.Subject
				if req.Attrs == consts.Pinned || req.Attrs == consts.PinnedAndHighlighted {
					data = &subjectMapper.Subject{ID: oid, UserId: req.UserId, TopCommentId: lo.ToPtr(req.Id)}
				} else {
					data = &subjectMapper.Subject{ID: oid, UserId: req.UserId, TopCommentId: lo.ToPtr("")}
				}
				if _, err = s.SubjectMongoMapper.Update(sessionContext, data); err != nil {
					if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
						log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)
						return err
					}
				}
			}

			data := convertor.CommentToCommentMapper(&gencomment.Comment{Id: req.Id, UserId: req.UserId, Attrs: req.Attrs, CreateTime: req.SortTime})
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
	}
	return resp, err
}
