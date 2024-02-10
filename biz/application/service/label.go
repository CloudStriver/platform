package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pagination/esp"
	"github.com/CloudStriver/go-pkg/utils/pagination/mongop"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/convertor"
	labelMapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/label"
	entityMapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/labelEntity"
	gencomment "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/comment"
	"github.com/google/wire"
	"github.com/samber/lo"
)

type ILabelService interface {
	CreateLabel(ctx context.Context, req *gencomment.CreateLabelReq) (resp *gencomment.CreateLabelResp, err error)
	DeleteLabel(ctx context.Context, req *gencomment.DeleteLabelReq) (resp *gencomment.DeleteLabelResp, err error)
	GetLabel(ctx context.Context, req *gencomment.GetLabelReq) (resp *gencomment.GetLabelResp, err error)
	GetLabelsInBatch(ctx context.Context, req *gencomment.GetLabelsInBatchReq) (resp *gencomment.GetLabelsInBatchResp, err error)
	UpdateLabel(ctx context.Context, req *gencomment.UpdateLabelReq) (resp *gencomment.UpdateLabelResp, err error)
	GetLabels(ctx context.Context, req *gencomment.GetLabelsReq) (resp *gencomment.GetLabelsResp, err error)
	CreateObject(ctx context.Context, req *gencomment.CreateObjectReq) (resp *gencomment.CreateObjectResp, err error)
	CreateObjects(ctx context.Context, req *gencomment.CreateObjectsReq) (resp *gencomment.CreateObjectsResp, err error)
	DeleteObject(ctx context.Context, req *gencomment.DeleteObjectReq) (resp *gencomment.DeleteObjectResp, err error)
	GetObjects(ctx context.Context, req *gencomment.GetObjectsReq) (resp *gencomment.GetObjectsResp, err error)
	UpdateObject(ctx context.Context, req *gencomment.UpdateObjectReq) (resp *gencomment.UpdateObjectResp, err error)
}

type LabelService struct {
	LabelEsMapper     labelMapper.IEsMapper
	LabelMongoMapper  labelMapper.IMongoMapper
	EntityMongoMapper entityMapper.IMongoMapper
}

var LabelSet = wire.NewSet(
	wire.Struct(new(LabelService), "*"),
	wire.Bind(new(ILabelService), new(*LabelService)),
)

func (s *LabelService) CreateLabel(ctx context.Context, req *gencomment.CreateLabelReq) (resp *gencomment.CreateLabelResp, err error) {
	resp = new(gencomment.CreateLabelResp)
	var id string
	if id, err = s.LabelMongoMapper.Insert(ctx, convertor.LabelToLabelMapper(req.Label)); err != nil {
		log.CtxError(ctx, "创建标签 失败[%v]\n", err)
		return resp, err
	}
	resp.Id = id
	return resp, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, req *gencomment.DeleteLabelReq) (resp *gencomment.DeleteLabelResp, err error) {
	resp = new(gencomment.DeleteLabelResp)
	if _, err = s.LabelMongoMapper.Delete(ctx, req.Id); err != nil {
		log.CtxError(ctx, "删除标签 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) GetLabel(ctx context.Context, req *gencomment.GetLabelReq) (resp *gencomment.GetLabelResp, err error) {
	resp = new(gencomment.GetLabelResp)
	var label *labelMapper.Label
	if label, err = s.LabelMongoMapper.FindOne(ctx, req.Id); err != nil {
		log.CtxError(ctx, "获取标签 失败[%v]\n", err)
		return resp, err
	}
	resp.Label = convertor.LabelMapperToLabel(label)
	return resp, nil
}

func (s *LabelService) UpdateLabel(ctx context.Context, req *gencomment.UpdateLabelReq) (resp *gencomment.UpdateLabelResp, err error) {
	resp = new(gencomment.UpdateLabelResp)
	if _, err = s.LabelMongoMapper.Update(ctx, convertor.LabelToLabelMapper(req.Label)); err != nil {
		log.CtxError(ctx, "获取标签 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) GetLabels(ctx context.Context, req *gencomment.GetLabelsReq) (resp *gencomment.GetLabelsResp, err error) {
	resp = new(gencomment.GetLabelsResp)
	var total int64
	var labels []*labelMapper.Label
	p := convertor.ParsePagination(req.Pagination)
	if labels, total, err = s.LabelEsMapper.Search(ctx, convertor.ConvertLabelAllFieldsSearchQuery(req.Key), p, esp.ScoreCursorType); err != nil {
		log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
		return resp, err
	}
	if p.LastToken != nil {
		resp.Token = *p.LastToken
	}
	resp.Labels = lo.Map(labels, func(item *labelMapper.Label, _ int) *gencomment.Label {
		return convertor.LabelMapperToLabel(item)
	})
	resp.Total = total
	return
}

func (s *LabelService) GetLabelsInBatch(ctx context.Context, req *gencomment.GetLabelsInBatchReq) (resp *gencomment.GetLabelsInBatchResp, err error) {
	resp = new(gencomment.GetLabelsInBatchResp)
	var labels []*labelMapper.Label
	if labels, err = s.LabelMongoMapper.FindManyByIds(ctx, req.LabelIds); err != nil {
		log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
		return resp, err
	}

	// 创建映射：标签ID到标签对象
	labelMap := make(map[string]*labelMapper.Label)
	for _, label := range labels {
		labelMap[label.ID.Hex()] = label
	}

	// 按req.LabelIds中的ID顺序映射和转换
	resp.Labels = lo.Map(req.LabelIds, func(id string, _ int) *gencomment.Label {
		if label, ok := labelMap[id]; ok {
			return convertor.LabelMapperToLabel(label)
		}
		return nil // 或者处理找不到标签的情况
	})
	return resp, nil
}

func (s *LabelService) CreateObject(ctx context.Context, req *gencomment.CreateObjectReq) (resp *gencomment.CreateObjectResp, err error) {
	resp = new(gencomment.CreateObjectResp)
	if err = s.EntityMongoMapper.Insert(ctx, convertor.LabelEntityToLabelEntityMapper(req.Object)); err != nil {
		log.CtxError(ctx, "创建标签实体 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) CreateObjects(ctx context.Context, req *gencomment.CreateObjectsReq) (resp *gencomment.CreateObjectsResp, err error) {
	resp = new(gencomment.CreateObjectsResp)
	data := lo.Map(req.Objects, func(item *gencomment.LabelEntity, _ int) *entityMapper.LabelEntity {
		return convertor.LabelEntityToLabelEntityMapper(item)
	})
	if err = s.EntityMongoMapper.InsertMany(ctx, data); err != nil {
		log.CtxError(ctx, "批量创建标签实体 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) DeleteObject(ctx context.Context, req *gencomment.DeleteObjectReq) (resp *gencomment.DeleteObjectResp, err error) {
	resp = new(gencomment.DeleteObjectResp)
	if _, err = s.EntityMongoMapper.Delete(ctx, req.ObjectId); err != nil {
		log.CtxError(ctx, "删除标签实体 失败[%v]\n", err)
		return resp, err
	}
	return
}
func (s *LabelService) GetObjects(ctx context.Context, req *gencomment.GetObjectsReq) (resp *gencomment.GetObjectsResp, err error) {
	resp = new(gencomment.GetObjectsResp)
	var total int64
	var objects []*entityMapper.LabelEntity
	p := convertor.ParsePagination(req.Pagination)
	filter := convertor.LabelEntityFilterOptionsToFilterOptions(req.FilterOptions)
	if objects, total, err = s.EntityMongoMapper.FindManyAndCount(ctx, filter, p, mongop.IdCursorType); err != nil {
		log.CtxError(ctx, "获取标签实体集 失败[%v]\n", err)
		return resp, err
	}

	if p.LastToken != nil {
		resp.Token = *p.LastToken
	}
	resp.ObjectIds = lo.Map(objects, func(item *entityMapper.LabelEntity, _ int) string {
		return item.ID.Hex()
	})
	resp.Total = total
	return resp, nil
}

func (s *LabelService) UpdateObject(ctx context.Context, req *gencomment.UpdateObjectReq) (resp *gencomment.UpdateObjectResp, err error) {
	resp = new(gencomment.UpdateObjectResp)
	if _, err = s.EntityMongoMapper.Update(ctx, convertor.LabelEntityToLabelEntityMapper(req.Object)); err != nil {
		log.CtxError(ctx, "更新标签实体 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}
