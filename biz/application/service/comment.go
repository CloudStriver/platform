package service

import (
	"context"
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
	UpdateCount(ctx context.Context, rootId, subjectId, fatherId string, count int64)
	GetComment(ctx context.Context, req *platform.GetCommentReq) (resp *platform.GetCommentResp, err error)
	GetCommentList(ctx context.Context, req *platform.GetCommentListReq) (resp *platform.GetCommentListResp, err error)
	CreateComment(ctx context.Context, req *platform.CreateCommentReq) (resp *platform.CreateCommentResp, err error)
	UpdateComment(ctx context.Context, req *platform.UpdateCommentReq) (resp *platform.UpdateCommentResp, err error)
	DeleteComment(ctx context.Context, req *platform.DeleteCommentReq) (resp *platform.DeleteCommentResp, err error)
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
	if _, err = s.CommentMongoMapper.DeleteMany(ctx, req.Ids); err != nil {
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
	resp.Comment = convertor.CommentMapperToCommentInfo(data)
	return resp, nil
}

func (s *CommentService) GetCommentList(ctx context.Context, req *platform.GetCommentListReq) (resp *platform.GetCommentListResp, err error) {
	resp = new(platform.GetCommentListResp)
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
	resp.Comments = lo.Map(comments, func(comment *commentMapper.Comment, _ int) *platform.CommentInfo {
		return convertor.CommentMapperToCommentInfo(comment)
	})
	resp.Total = total
	return resp, nil
}

func (s *CommentService) CreateComment(ctx context.Context, req *platform.CreateCommentReq) (resp *platform.CreateCommentResp, err error) {
	resp = new(platform.CreateCommentResp)
	data := convertor.CommentToCommentMapper(req.Comment)
	if resp.Id, err = s.CommentMongoMapper.Insert(ctx, data); err != nil {
		log.CtxError(ctx, "创建评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) UpdateCount(ctx context.Context, rootId, subjectId, fatherId string, count int64) {
	if rootId != subjectId {
		if fatherId != subjectId {
			// 二级评论 + 三级评论
			s.CommentMongoMapper.UpdateCount(ctx, rootId, count)
		}
	}
}

func (s *CommentService) UpdateComment(ctx context.Context, req *platform.UpdateCommentReq) (resp *platform.UpdateCommentResp, err error) {
	resp = new(platform.UpdateCommentResp)
	data := convertor.CommentToCommentMapper(req.Comment)
	if _, err = s.CommentMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "更新评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *platform.DeleteCommentReq) (resp *platform.DeleteCommentResp, err error) {
	resp = new(platform.DeleteCommentResp)
	if _, err = s.CommentMongoMapper.Delete(ctx, req.Id); err != nil {
		log.CtxError(ctx, "删除评论 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *CommentService) SetCommentAttrs(ctx context.Context, req *platform.SetCommentAttrsReq, res *platform.GetCommentSubjectResp) (resp *platform.SetCommentAttrsResp, err error) {
	resp = new(platform.SetCommentAttrsResp)
	if req.Attrs == int64(platform.Attrs_Pinned) || req.Attrs == int64(platform.Attrs_PinnedAndHighlighted) {
		req.SortTime = math.MaxInt64 - 1
	}

	oid, _ := primitive.ObjectIDFromHex(req.SubjectId)
	data := convertor.CommentToCommentMapper(&platform.Comment{Id: req.Id, Attrs: req.Attrs, SortTime: req.SortTime})
	tx := s.SubjectMongoMapper.StartClient()
	if res.Subject.TopCommentId == "" {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == int64(platform.Attrs_Pinned) || req.Attrs == int64(platform.Attrs_PinnedAndHighlighted) {
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: oid, TopCommentId: lo.ToPtr(req.Id)}); err != nil {
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
	} else if res.Subject.TopCommentId == req.Id {
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			if req.Attrs == int64(platform.Attrs_None) || req.Attrs == int64(platform.Attrs_Highlighted) {
				if _, err = s.SubjectMongoMapper.Update(sessionContext, &subjectMapper.Subject{ID: oid, TopCommentId: lo.ToPtr("")}); err != nil {
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
		if oldComment, err = s.CommentMongoMapper.FindOne(ctx, res.Subject.TopCommentId); err != nil {
			return resp, err
		}
		err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
			if err = sessionContext.StartTransaction(); err != nil {
				return err
			}
			var subject *subjectMapper.Subject
			switch {
			case req.Attrs == int64(platform.Attrs_Pinned) || req.Attrs == int64(platform.Attrs_PinnedAndHighlighted):
				subject = &subjectMapper.Subject{ID: oid, TopCommentId: lo.ToPtr(req.Id)}
			default:
				subject = &subjectMapper.Subject{ID: oid, TopCommentId: lo.ToPtr("")}
			}
			if _, err = s.SubjectMongoMapper.Update(sessionContext, subject); err != nil {
				if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
					log.CtxError(sessionContext, "设置评论属性失败[%v]: 回滚异常[%v]\n", err, rbErr)

					return err
				}
			}
			oldData := convertor.CommentToCommentMapper(&platform.Comment{Id: res.Subject.TopCommentId, Attrs: int64(platform.Attrs_None), SortTime: oldComment.CreateAt.UnixMilli()})
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
