package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	relationmapper "github.com/CloudStriver/platform-comment/biz/infrastructure/mapper/relation"
	genrelation "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/relation"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RelationService interface {
	CreateRelation(ctx context.Context, req *genrelation.CreateRelationReq) (resp *genrelation.CreateRelationResp, err error)
	GetRelation(ctx context.Context, req *genrelation.GetRelationReq) (resp *genrelation.GetRelationResp, err error)
	DeleteRelation(ctx context.Context, req *genrelation.DeleteRelationReq) (resp *genrelation.DeleteRelationResp, err error)
	GetRelations(ctx context.Context, req *genrelation.GetRelationsReq) (resp *genrelation.GetRelationsResp, err error)
	GetRelationCount(ctx context.Context, req *genrelation.GetRelationCountReq) (resp *genrelation.GetRelationCountResp, err error)
	GetRelationPaths(ctx context.Context, req *genrelation.GetRelationPathsReq) (resp *genrelation.GetRelationPathsResp, err error)
	DeleteNode(ctx context.Context, req *genrelation.DeleteNodeReq) (resp *genrelation.DeleteNodeResp, err error)
}

var RelationSet = wire.NewSet(
	wire.Struct(new(RelationServiceImpl), "*"),
	wire.Bind(new(RelationService), new(*RelationServiceImpl)),
)

type RelationServiceImpl struct {
	Config        *config.Config
	Redis         *redis.Redis
	RelationModel relationmapper.RelationNeo4jMapper
}

func (s *RelationServiceImpl) DeleteNode(ctx context.Context, req *genrelation.DeleteNodeReq) (resp *genrelation.DeleteNodeResp, err error) {
	resp = new(genrelation.DeleteNodeResp)
	if err = s.RelationModel.DeleteNode(ctx, req.NodeId, req.NodeType); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelationPaths(ctx context.Context, req *genrelation.GetRelationPathsReq) (resp *genrelation.GetRelationPathsResp, err error) {
	resp = new(genrelation.GetRelationPathsResp)
	p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
	resp.Relations, err = s.RelationModel.GetRelationPaths(ctx, req.FromType, req.FromId, req.EdgeType1, req.EdgeType2, p)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelationCount(ctx context.Context, req *genrelation.GetRelationCountReq) (resp *genrelation.GetRelationCountResp, err error) {
	resp = new(genrelation.GetRelationCountResp)
	switch o := req.RelationFilterOptions.(type) {
	case *genrelation.GetRelationCountReq_FromFilterOptions:
		resp.Total, err = s.RelationModel.MatchFromEdgesCount(ctx, o.FromFilterOptions.FromType, o.FromFilterOptions.FromId, o.FromFilterOptions.ToType, req.RelationType)
	case *genrelation.GetRelationCountReq_ToFilterOptions:
		resp.Total, err = s.RelationModel.MatchToEdgesCount(ctx, o.ToFilterOptions.ToType, o.ToFilterOptions.ToId, o.ToFilterOptions.FromType, req.RelationType)
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelations(ctx context.Context, req *genrelation.GetRelationsReq) (resp *genrelation.GetRelationsResp, err error) {
	resp = new(genrelation.GetRelationsResp)
	p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
	switch o := req.RelationFilterOptions.(type) {
	case *genrelation.GetRelationsReq_FromFilterOptions:
		resp.Relations, resp.Total, err = s.RelationModel.MatchFromEdgesAndCount(ctx, o.FromFilterOptions.FromType, o.FromFilterOptions.FromId, o.FromFilterOptions.ToType,
			req.RelationType, p)
	case *genrelation.GetRelationsReq_ToFilterOptions:
		resp.Relations, resp.Total, err = s.RelationModel.MatchToEdgesAndCount(ctx, o.ToFilterOptions.ToType, o.ToFilterOptions.ToId, o.ToFilterOptions.FromType,
			req.RelationType, p)
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) DeleteRelation(ctx context.Context, req *genrelation.DeleteRelationReq) (resp *genrelation.DeleteRelationResp, err error) {
	if err = s.RelationModel.DeleteEdge(ctx, &genrelation.Relation{
		FromType:     req.FromType,
		FromId:       req.FromId,
		ToType:       req.ToType,
		ToId:         req.ToId,
		RelationType: req.RelationType,
	}); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) CreateRelation(ctx context.Context, req *genrelation.CreateRelationReq) (resp *genrelation.CreateRelationResp, err error) {
	resp = new(genrelation.CreateRelationResp)
	ok, err := s.RelationModel.MatchEdge(ctx, &genrelation.Relation{
		FromType:     req.FromType,
		FromId:       req.FromId,
		ToType:       req.ToType,
		ToId:         req.ToId,
		RelationType: req.RelationType,
	})
	if err != nil {
		return resp, err
	}
	if !ok {
		if err = s.RelationModel.CreateEdge(ctx, &genrelation.Relation{
			FromType:     req.FromType,
			FromId:       req.FromId,
			ToType:       req.ToType,
			ToId:         req.ToId,
			RelationType: req.RelationType,
		}); err != nil {
			return resp, err
		}
		resp.Ok = true
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelation(ctx context.Context, req *genrelation.GetRelationReq) (resp *genrelation.GetRelationResp, err error) {
	resp = new(genrelation.GetRelationResp)
	if resp.Ok, err = s.RelationModel.MatchEdge(ctx, &genrelation.Relation{
		FromType:     req.FromType,
		FromId:       req.FromId,
		ToType:       req.ToType,
		ToId:         req.ToId,
		RelationType: req.RelationType,
	}); err != nil {
		return resp, err
	}
	return resp, nil
}
