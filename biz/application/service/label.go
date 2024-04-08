package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pagination/esp"
	"github.com/CloudStriver/go-pkg/utils/pagination/mongop"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform/biz/infrastructure/convertor"
	labelMapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/label"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/google/wire"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ILabelService interface {
	CreateLabel(ctx context.Context, req *platform.CreateLabelReq) (resp *platform.CreateLabelResp, err error)
	DeleteLabel(ctx context.Context, req *platform.DeleteLabelReq) (resp *platform.DeleteLabelResp, err error)
	GetLabel(ctx context.Context, req *platform.GetLabelReq) (resp *platform.GetLabelResp, err error)
	GetLabelsInBatch(ctx context.Context, req *platform.GetLabelsInBatchReq) (resp *platform.GetLabelsInBatchResp, err error)
	UpdateLabel(ctx context.Context, req *platform.UpdateLabelReq) (resp *platform.UpdateLabelResp, err error)
	GetLabels(ctx context.Context, req *platform.GetLabelsReq) (resp *platform.GetLabelsResp, err error)
}

type LabelService struct {
	LabelEsMapper    labelMapper.IEsMapper
	LabelMongoMapper labelMapper.IMongoMapper
}

var LabelSet = wire.NewSet(
	wire.Struct(new(LabelService), "*"),
	wire.Bind(new(ILabelService), new(*LabelService)),
)

func (s *LabelService) CreateLabel(ctx context.Context, req *platform.CreateLabelReq) (resp *platform.CreateLabelResp, err error) {
	resp = new(platform.CreateLabelResp)
	var id string
	if id, err = s.LabelMongoMapper.Insert(ctx, &labelMapper.Label{
		ID:       primitive.NilObjectID,
		FatherId: req.FatherId,
		Value:    req.Value,
	}); err != nil {
		log.CtxError(ctx, "创建标签 失败[%v]\n", err)
		return resp, err
	}
	resp.Id = id
	return resp, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, req *platform.DeleteLabelReq) (resp *platform.DeleteLabelResp, err error) {
	resp = new(platform.DeleteLabelResp)
	if _, err = s.LabelMongoMapper.Delete(ctx, req.Id); err != nil {
		log.CtxError(ctx, "删除标签 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) GetLabel(ctx context.Context, req *platform.GetLabelReq) (resp *platform.GetLabelResp, err error) {
	resp = new(platform.GetLabelResp)
	var label *labelMapper.Label
	if label, err = s.LabelMongoMapper.FindOne(ctx, req.Id); err != nil {
		log.CtxError(ctx, "获取标签 失败[%v]\n", err)
		return resp, err
	}
	resp.Value = label.Value
	return resp, nil
}

func (s *LabelService) UpdateLabel(ctx context.Context, req *platform.UpdateLabelReq) (resp *platform.UpdateLabelResp, err error) {
	resp = new(platform.UpdateLabelResp)
	var oid primitive.ObjectID
	if oid, err = primitive.ObjectIDFromHex(req.Id); err != nil {
		return resp, err
	}
	if _, err = s.LabelMongoMapper.Update(ctx, &labelMapper.Label{
		ID:       oid,
		FatherId: req.FatherId,
		Value:    req.Value,
	}); err != nil {
		log.CtxError(ctx, "获取标签 失败[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *LabelService) GetLabels(ctx context.Context, req *platform.GetLabelsReq) (resp *platform.GetLabelsResp, err error) {
	resp = new(platform.GetLabelsResp)
	var total int64
	var labels []*labelMapper.Label
	p := convertor.ParsePagination(req.Pagination)

	if req.Key != nil {
		switch {
		case *req.Key == "":
			fopts := &labelMapper.FilterOptions{OnlyFatherId: req.FatherId}
			if labels, total, err = s.LabelMongoMapper.FindManyAndCount(ctx, fopts, p, mongop.IdCursorType); err != nil {
				log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
				return resp, err
			}
		case *req.Key != "":
			if labels, total, err = s.LabelEsMapper.Search(ctx, convertor.ConvertLabelAllFieldsSearchQuery(*req.Key), p, esp.ScoreCursorType); err != nil {
				log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
				return resp, err
			}
		}
	} else {
		fopts := &labelMapper.FilterOptions{OnlyFatherId: req.FatherId}
		if labels, total, err = s.LabelMongoMapper.FindManyAndCount(ctx, fopts, p, mongop.IdCursorType); err != nil {
			log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
			return resp, err
		}
	}

	if p.LastToken != nil {
		resp.Token = *p.LastToken
	}
	resp.Labels = lo.Map(labels, func(item *labelMapper.Label, _ int) *platform.Label {
		return convertor.LabelMapperToLabel(item)
	})
	resp.Total = total
	return
}

func (s *LabelService) GetLabelsInBatch(ctx context.Context, req *platform.GetLabelsInBatchReq) (resp *platform.GetLabelsInBatchResp, err error) {
	resp = new(platform.GetLabelsInBatchResp)
	var labels []*labelMapper.Label
	if labels, err = s.LabelMongoMapper.FindManyByIds(ctx, req.Ids); err != nil {
		log.CtxError(ctx, "获取标签集 失败[%v]\n", err)
		return resp, err
	}

	// 创建映射：标签ID到标签对象
	labelMap := make(map[string]string)
	for _, label := range labels {
		labelMap[label.ID.Hex()] = label.Value
	}

	// 按req.LabelIds中的ID顺序映射和转换
	resp.Labels = lo.Map(req.Ids, func(id string, _ int) string {
		if label, ok := labelMap[id]; ok {
			return label
		}
		return "" // 或者处理找不到标签的情况
	})
	return resp, nil
}
