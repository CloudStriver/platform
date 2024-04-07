package service

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/platform/biz/infrastructure/config"
	relationmapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/relation"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RelationService interface {
	CreateRelation(ctx context.Context, req *platform.CreateRelationReq) (resp *platform.CreateRelationResp, err error)
	GetRelation(ctx context.Context, req *platform.GetRelationReq) (resp *platform.GetRelationResp, err error)
	DeleteRelation(ctx context.Context, req *platform.DeleteRelationReq) (resp *platform.DeleteRelationResp, err error)
	GetRelations(ctx context.Context, req *platform.GetRelationsReq) (resp *platform.GetRelationsResp, err error)
	GetRelationCount(ctx context.Context, req *platform.GetRelationCountReq) (resp *platform.GetRelationCountResp, err error)
	GetRelationPaths(ctx context.Context, req *platform.GetRelationPathsReq) (resp *platform.GetRelationPathsResp, err error)
	DeleteNode(ctx context.Context, req *platform.DeleteNodeReq) (resp *platform.DeleteNodeResp, err error)
	GetRelationPathsCount(ctx context.Context, req *platform.GetRelationPathsCountReq) (resp *platform.GetRelationPathsCountResp, err error)
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

func (s *RelationServiceImpl) GetRelationPathsCount(ctx context.Context, req *platform.GetRelationPathsCountReq) (resp *platform.GetRelationPathsCountResp, err error) {
	resp = new(platform.GetRelationPathsCountResp)
	resp.Total, err = s.RelationModel.GetRelationPathsCount(ctx, req.FromType1, req.FromId1, req.FromType2, req.FromId2, req.EdgeType1, req.EdgeType2, req.ToType)
	if err != nil {
		return resp, err
	}
	return resp, nil

}

func (s *RelationServiceImpl) DeleteNode(ctx context.Context, req *platform.DeleteNodeReq) (resp *platform.DeleteNodeResp, err error) {
	resp = new(platform.DeleteNodeResp)
	if err = s.RelationModel.DeleteNode(ctx, req.NodeId, req.NodeType); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelationPaths(ctx context.Context, req *platform.GetRelationPathsReq) (resp *platform.GetRelationPathsResp, err error) {
	resp = new(platform.GetRelationPathsResp)
	p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
	resp.Relations, err = s.RelationModel.GetRelationPaths(ctx, req.FromType, req.FromId, req.EdgeType1, req.EdgeType2, p)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelationCount(ctx context.Context, req *platform.GetRelationCountReq) (resp *platform.GetRelationCountResp, err error) {
	resp = new(platform.GetRelationCountResp)
	switch o := req.RelationFilterOptions.(type) {
	case *platform.GetRelationCountReq_FromFilterOptions:
		resp.Total, err = s.RelationModel.MatchFromEdgesCount(ctx, o.FromFilterOptions.FromType, o.FromFilterOptions.FromId, o.FromFilterOptions.ToType, req.RelationType)
	case *platform.GetRelationCountReq_ToFilterOptions:
		resp.Total, err = s.RelationModel.MatchToEdgesCount(ctx, o.ToFilterOptions.ToType, o.ToFilterOptions.ToId, o.ToFilterOptions.FromType, req.RelationType)
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelations(ctx context.Context, req *platform.GetRelationsReq) (resp *platform.GetRelationsResp, err error) {
	resp = new(platform.GetRelationsResp)
	p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
	switch o := req.RelationFilterOptions.(type) {
	case *platform.GetRelationsReq_FromFilterOptions:
		resp.Relations, resp.Total, err = s.RelationModel.MatchFromEdgesAndCount(ctx, o.FromFilterOptions.FromType, o.FromFilterOptions.FromId, o.FromFilterOptions.ToType,
			req.RelationType, p)
	case *platform.GetRelationsReq_ToFilterOptions:
		resp.Relations, resp.Total, err = s.RelationModel.MatchToEdgesAndCount(ctx, o.ToFilterOptions.ToType, o.ToFilterOptions.ToId, o.ToFilterOptions.FromType,
			req.RelationType, p)
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) DeleteRelation(ctx context.Context, req *platform.DeleteRelationReq) (resp *platform.DeleteRelationResp, err error) {
	if err = s.RelationModel.DeleteEdge(ctx, &platform.Relation{
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

func (s *RelationServiceImpl) CreateRelation(ctx context.Context, req *platform.CreateRelationReq) (resp *platform.CreateRelationResp, err error) {
	resp = new(platform.CreateRelationResp)
	ok, err := s.RelationModel.MatchEdge(ctx, &platform.Relation{
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
		if err = s.RelationModel.CreateEdge(ctx, &platform.Relation{
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

func (s *RelationServiceImpl) GetRelation(ctx context.Context, req *platform.GetRelationReq) (resp *platform.GetRelationResp, err error) {
	resp = new(platform.GetRelationResp)
	if resp.Ok, err = s.RelationModel.MatchEdge(ctx, &platform.Relation{
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
