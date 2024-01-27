package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/convertor"
	subjectMapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/subject"
	gencomment "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	"github.com/google/wire"
	"github.com/samber/lo"
)

type ISubjectService interface {
	UpdateAfterCreateComment(ctx context.Context, req *gencomment.Comment)
	GetCommentSubject(ctx context.Context, req *gencomment.GetCommentSubjectReq) (resp *gencomment.GetCommentSubjectResp, err error)
	CreateCommentSubject(ctx context.Context, req *gencomment.CreateCommentSubjectReq) (resp *gencomment.CreateCommentSubjectResp, err error)
	UpdateCommentSubject(ctx context.Context, req *gencomment.UpdateCommentSubjectReq) (resp *gencomment.UpdateCommentSubjectResp, err error)
	DeleteCommentSubject(ctx context.Context, req *gencomment.DeleteCommentSubjectReq) (resp *gencomment.DeleteCommentSubjectResp, err error)
	SetCommentSubjectState(ctx context.Context, req *gencomment.SetCommentSubjectStateReq) (resp *gencomment.SetCommentSubjectStateResp, err error)
	SetCommentSubjectAttrs(ctx context.Context, req *gencomment.SetCommentSubjectAttrsReq) (resp *gencomment.SetCommentSubjectAttrsResp, err error)
}

type SubjectService struct {
	SubjectMongoMapper subjectMapper.IMongoMapper
}

var SubjectSet = wire.NewSet(
	wire.Struct(new(SubjectService), "*"),
	wire.Bind(new(ISubjectService), new(*SubjectService)),
)

func (s *SubjectService) GetCommentSubject(ctx context.Context, req *gencomment.GetCommentSubjectReq) (resp *gencomment.GetCommentSubjectResp, err error) {
	resp = new(gencomment.GetCommentSubjectResp)
	var data *subjectMapper.Subject
	filter := convertor.SubjectFilterOptionsToFilterOptions(req.FilterOptions)
	if data, err = s.SubjectMongoMapper.FindOne(ctx, filter); err != nil {
		log.CtxError(ctx, "获取评论区详情 失败[%v]\n", err)
		return resp, err
	}
	resp.Subject = convertor.SubjectMapperToSubjectDetail(data)
	return resp, nil
}

func (s *SubjectService) CreateCommentSubject(ctx context.Context, req *gencomment.CreateCommentSubjectReq) (resp *gencomment.CreateCommentSubjectResp, err error) {
	resp = new(gencomment.CreateCommentSubjectResp)
	data := convertor.SubjectToSubjectMapper(req.Subject)
	if resp.Id, err = s.SubjectMongoMapper.Insert(ctx, data); err != nil {
		log.CtxError(ctx, "创建评论区 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) UpdateAfterCreateComment(ctx context.Context, req *gencomment.Comment) {
	if req.RootId == req.SubjectId {
		// 一级评论
		if req.FatherId == req.SubjectId {
			data := convertor.SubjectToSubjectMapper(&gencomment.Subject{Id: req.SubjectId, RootCount: lo.ToPtr(int64(1)), AllCount: lo.ToPtr(int64(1))})
			s.SubjectMongoMapper.UpdateAfterCreateComment(ctx, data)
		}
	} else {
		// 二级评论 + 三级评论
		if req.FatherId != req.SubjectId {
			data := convertor.SubjectToSubjectMapper(&gencomment.Subject{Id: req.SubjectId, AllCount: lo.ToPtr(int64(1))})
			s.SubjectMongoMapper.UpdateAfterCreateComment(ctx, data)
		}
	}
}

func (s *SubjectService) UpdateCommentSubject(ctx context.Context, req *gencomment.UpdateCommentSubjectReq) (resp *gencomment.UpdateCommentSubjectResp, err error) {
	resp = new(gencomment.UpdateCommentSubjectResp)
	data := convertor.SubjectToSubjectMapper(req.Subject)
	if _, err = s.SubjectMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "修改评论区信息 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) DeleteCommentSubject(ctx context.Context, req *gencomment.DeleteCommentSubjectReq) (resp *gencomment.DeleteCommentSubjectResp, err error) {
	resp = new(gencomment.DeleteCommentSubjectResp)
	if _, err = s.SubjectMongoMapper.Delete(ctx, req.Id, req.UserId); err != nil {
		log.CtxError(ctx, "删除评论区 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) SetCommentSubjectState(ctx context.Context, req *gencomment.SetCommentSubjectStateReq) (resp *gencomment.SetCommentSubjectStateResp, err error) {
	resp = new(gencomment.SetCommentSubjectStateResp)
	data := convertor.SubjectToSubjectMapper(&gencomment.Subject{Id: req.Id, UserId: req.UserId, State: req.State})
	if _, err = s.SubjectMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "设置评论区状态 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) SetCommentSubjectAttrs(ctx context.Context, req *gencomment.SetCommentSubjectAttrsReq) (resp *gencomment.SetCommentSubjectAttrsResp, err error) {
	resp = new(gencomment.SetCommentSubjectAttrsResp)
	data := convertor.SubjectToSubjectMapper(&gencomment.Subject{Id: req.Id, UserId: req.UserId, Attrs: req.Attrs})
	if _, err = s.SubjectMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "设置评论区属性 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}
