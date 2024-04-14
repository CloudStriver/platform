package service

import (
	"context"
	"github.com/CloudStriver/cloudmind-mq/app/util/message"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform/biz/infrastructure/kq"
	subjectMapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/subject"
	gencontent "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/content"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/bytedance/sonic"
	"github.com/google/wire"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ISubjectService interface {
	UpdateCount(ctx context.Context, subjectId string, rootCount, allCount int64)
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
	if data, err = s.SubjectMongoMapper.FindOne(ctx, req.SubjectId); err != nil {
		log.CtxError(ctx, "获取评论区详情 失败[%v]\n", err)
		return resp, err
	}

	resp = &platform.GetCommentSubjectResp{
		UserId:       data.UserId,
		TopCommentId: *data.TopCommentId,
		RootCount:    *data.RootCount,
		AllCount:     *data.AllCount,
		State:        data.State,
		Attrs:        data.Attrs,
	}
	return resp, nil
}

func (s *SubjectService) CreateCommentSubject(ctx context.Context, req *platform.CreateCommentSubjectReq) (resp *platform.CreateCommentSubjectResp, err error) {
	resp = new(platform.CreateCommentSubjectResp)
	var oid primitive.ObjectID
	if oid, err = primitive.ObjectIDFromHex(req.SubjectId); err != nil {
		return resp, err
	}
	if _, err = s.SubjectMongoMapper.Insert(ctx, &subjectMapper.Subject{
		ID:           oid,
		UserId:       req.UserId,
		TopCommentId: lo.ToPtr(""),
		RootCount:    lo.ToPtr(int64(0)),
		AllCount:     lo.ToPtr(int64(0)),
		State:        int64(platform.State_Normal),
		Attrs:        int64(platform.Attrs_None),
	}); err != nil {
		log.CtxError(ctx, "创建评论区 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) UpdateCount(ctx context.Context, subjectId string, rootCount, allCount int64) {
	s.SubjectMongoMapper.UpdateCount(ctx, subjectId, rootCount, allCount)
}

func (s *SubjectService) UpdateCommentSubject(ctx context.Context, req *platform.UpdateCommentSubjectReq) (resp *platform.UpdateCommentSubjectResp, err error) {
	resp = new(platform.UpdateCommentSubjectResp)
	var oid primitive.ObjectID
	if oid, err = primitive.ObjectIDFromHex(req.SubjectId); err != nil {
		return resp, err
	}
	if _, err = s.SubjectMongoMapper.Update(ctx, &subjectMapper.Subject{
		ID:           oid,
		TopCommentId: nil,
		RootCount:    req.RootCount,
		AllCount:     req.AllCount,
		State:        req.State,
		Attrs:        req.Attrs,
	}); err != nil {
		log.CtxError(ctx, "修改评论区信息 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *SubjectService) DeleteCommentSubject(ctx context.Context, req *platform.DeleteCommentSubjectReq) (resp *platform.DeleteCommentSubjectResp, err error) {
	resp = new(platform.DeleteCommentSubjectResp)
	if _, err = s.SubjectMongoMapper.Delete(ctx, req.SubjectId); err != nil {
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
