package service

import (
	"context"
	"github.com/CloudStriver/cloudmind-mq/app/util/message"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/CloudStriver/platform/biz/infrastructure/convertor"
	"github.com/CloudStriver/platform/biz/infrastructure/kq"
	subjectMapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/subject"
	gencontent "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/content"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/bytedance/sonic"
	"github.com/google/wire"
)

type ISubjectService interface {
	UpdateCount(ctx context.Context, rootId, subjectId, fatherId string, count int64)
	GetCommentSubject(ctx context.Context, req *platform.GetCommentSubjectReq) (resp *platform.GetCommentSubjectResp, err error)
	CreateCommentSubject(ctx context.Context, req *platform.CreateCommentSubjectReq) (resp *platform.CreateCommentSubjectResp, err error)
	UpdateCommentSubject(ctx context.Context, req *platform.UpdateCommentSubjectReq) (resp *platform.UpdateCommentSubjectResp, err error)
	DeleteCommentSubject(ctx context.Context, req *platform.DeleteCommentSubjectReq) (resp *platform.DeleteCommentSubjectResp, err error)
}

type SubjectService struct {
	SubjectMongoMapper   subjectMapper.IMongoMapper
	DeleteFileRelationKq *kq.DeleteCommentRelationKq
}

var SubjectSet = wire.NewSet(
	wire.Struct(new(SubjectService), "*"),
	wire.Bind(new(ISubjectService), new(*SubjectService)),
)

func (s *SubjectService) GetCommentSubject(ctx context.Context, req *platform.GetCommentSubjectReq) (resp *platform.GetCommentSubjectResp, err error) {
	resp = new(platform.GetCommentSubjectResp)
	var data *subjectMapper.Subject
	if data, err = s.SubjectMongoMapper.FindOne(ctx, req.Id); err != nil {
		log.CtxError(ctx, "获取评论区详情 失败[%v]\n", err)
		return resp, err
	}
	resp.Subject = convertor.SubjectMapperToSubjectDetail(data)
	return resp, nil
}

func (s *SubjectService) CreateCommentSubject(ctx context.Context, req *platform.CreateCommentSubjectReq) (resp *platform.CreateCommentSubjectResp, err error) {
	resp = new(platform.CreateCommentSubjectResp)
	data := convertor.SubjectToSubjectMapper(req.Subject)
	if resp.Id, err = s.SubjectMongoMapper.Insert(ctx, data); err != nil {
		log.CtxError(ctx, "创建评论区 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) UpdateCount(ctx context.Context, rootId, subjectId, fatherId string, count int64) {
	if rootId == subjectId {
		// 一级评论
		if fatherId == subjectId {
			s.SubjectMongoMapper.UpdateCount(ctx, subjectId, count, count)
		}
	} else {
		// 二级评论 + 三级评论
		if fatherId != subjectId {
			s.SubjectMongoMapper.UpdateCount(ctx, subjectId, count, consts.InitNumber)
		}
	}
}

func (s *SubjectService) UpdateCommentSubject(ctx context.Context, req *platform.UpdateCommentSubjectReq) (resp *platform.UpdateCommentSubjectResp, err error) {
	resp = new(platform.UpdateCommentSubjectResp)
	data := convertor.SubjectToSubjectMapper(req.Subject)
	if _, err = s.SubjectMongoMapper.Update(ctx, data); err != nil {
		log.CtxError(ctx, "修改评论区信息 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) DeleteCommentSubject(ctx context.Context, req *platform.DeleteCommentSubjectReq) (resp *platform.DeleteCommentSubjectResp, err error) {
	resp = new(platform.DeleteCommentSubjectResp)
	if _, err = s.SubjectMongoMapper.Delete(ctx, req.Id); err != nil {
		log.CtxError(ctx, "删除评论区 失败[%v]\n", err)
		return resp, err
	}
	// 发送删除评论区关联文件的消息
	data, _ := sonic.Marshal(&message.DeleteCommentRelationsMessage{
		FromType: int64(gencontent.TargetType_UserType),
		FromId:   req.UserId,
	})
	if err2 := s.DeleteFileRelationKq.Push(pconvertor.Bytes2String(data)); err2 != nil {
		return resp, err2
	}
	return resp, nil
}
