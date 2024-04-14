package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform/biz/infrastructure/config"
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/CloudStriver/platform/biz/infrastructure/convertor"
	relationmapper "github.com/CloudStriver/platform/biz/infrastructure/mapper/relation"
	"github.com/CloudStriver/platform/biz/infrastructure/sort"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform"
	"github.com/google/wire"
	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	Config              *config.Config
	Redis               *redis.Redis
	RelationModel       relationmapper.RelationNeo4jMapper
	RelationMongoMapper relationmapper.IMongoMapper
}

func (s *RelationServiceImpl) GetRelationPathsCount(ctx context.Context, req *platform.GetRelationPathsCountReq) (resp *platform.GetRelationPathsCountResp, err error) {
	resp = new(platform.GetRelationPathsCountResp)
	if s.Config.Neo4jConf.Enable {
		resp.Total, err = s.RelationModel.GetRelationPathsCount(ctx, req.FromType1, req.FromId1, req.FromType2, req.EdgeType1, req.EdgeType2, req.ToType)
		if err != nil {
			return resp, err
		}
	} else {
		return resp, consts.ErrComponentNotStarted
	}
	return resp, nil

}

func (s *RelationServiceImpl) DeleteNode(ctx context.Context, req *platform.DeleteNodeReq) (resp *platform.DeleteNodeResp, err error) {
	resp = new(platform.DeleteNodeResp)

	tx := s.RelationMongoMapper.StartClient()
	if err = tx.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		var err1 error
		if err1 = sessionContext.StartTransaction(); err1 != nil {
			return err1
		}
		if _, err1 = s.RelationMongoMapper.Delete(sessionContext, &relationmapper.FilterOptions{
			OnlyFromType: lo.ToPtr(req.NodeType),
			OnlyFromId:   lo.ToPtr(req.NodeId),
		}); err1 != nil {
			if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
				log.CtxError(sessionContext, "保存文件中产生错误[%v]: 回滚异常[%v]\n", err1, rbErr)
			}
			return err1
		}

		if _, err1 = s.RelationMongoMapper.Delete(sessionContext, &relationmapper.FilterOptions{
			OnlyToType: lo.ToPtr(req.NodeType),
			OnlyToId:   lo.ToPtr(req.NodeId),
		}); err1 != nil {
			if rbErr := sessionContext.AbortTransaction(sessionContext); rbErr != nil {
				log.CtxError(sessionContext, "保存文件中产生错误[%v]: 回滚异常[%v]\n", err1, rbErr)
			}
			return err1
		}

		if err1 = sessionContext.CommitTransaction(sessionContext); err1 != nil {
			log.CtxError(sessionContext, "保存文件: 提交事务异常[%v]\n", err1)
			return err1
		}
		return nil
	}); err != nil {
		return resp, err
	}

	if s.Config.Neo4jConf.Enable {
		if err = s.RelationModel.DeleteNode(ctx, req.NodeId, req.NodeType); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func (s *RelationServiceImpl) GetRelationPaths(ctx context.Context, req *platform.GetRelationPathsReq) (resp *platform.GetRelationPathsResp, err error) {
	resp = new(platform.GetRelationPathsResp)
	if s.Config.Neo4jConf.Enable {
		p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
		if resp.Relations, err = s.RelationModel.GetRelationPaths(ctx, req.FromType, req.FromId, req.EdgeType1, req.EdgeType2, p); err != nil {
			return resp, err
		}
	} else {
		return resp, consts.ErrComponentNotStarted
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelationCount(ctx context.Context, req *platform.GetRelationCountReq) (resp *platform.GetRelationCountResp, err error) {
	resp = new(platform.GetRelationCountResp)
	switch o := req.RelationFilterOptions.(type) {
	case *platform.GetRelationCountReq_FromFilterOptions:
		if _, resp.Total, err = s.RelationMongoMapper.FindManyAndCount(ctx, &relationmapper.FilterOptions{
			OnlyFromType:     lo.ToPtr(o.FromFilterOptions.FromType),
			OnlyFromId:       lo.ToPtr(o.FromFilterOptions.FromId),
			OnlyToType:       lo.ToPtr(o.FromFilterOptions.ToType),
			OnlyRelationType: lo.ToPtr(req.RelationType),
		}, &pagination.PaginationOptions{}, sort.TimeCursorType); err != nil {
			return resp, err
		}
	case *platform.GetRelationCountReq_ToFilterOptions:
		if _, resp.Total, err = s.RelationMongoMapper.FindManyAndCount(ctx, &relationmapper.FilterOptions{
			OnlyFromType:     lo.ToPtr(o.ToFilterOptions.FromType),
			OnlyToId:         lo.ToPtr(o.ToFilterOptions.ToId),
			OnlyToType:       lo.ToPtr(o.ToFilterOptions.ToType),
			OnlyRelationType: lo.ToPtr(req.RelationType),
		}, &pagination.PaginationOptions{}, sort.TimeCursorType); err != nil {
			return resp, err
		}
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelations(ctx context.Context, req *platform.GetRelationsReq) (resp *platform.GetRelationsResp, err error) {
	resp = new(platform.GetRelationsResp)

	var (
		total     int64
		relations []*relationmapper.Relation
	)

	p := pconvertor.PaginationOptionsToModelPaginationOptions(req.PaginationOptions)
	switch o := req.RelationFilterOptions.(type) {
	case *platform.GetRelationsReq_FromFilterOptions:
		if relations, total, err = s.RelationMongoMapper.FindManyAndCount(ctx, &relationmapper.FilterOptions{
			OnlyFromType:     lo.ToPtr(o.FromFilterOptions.FromType),
			OnlyFromId:       lo.ToPtr(o.FromFilterOptions.FromId),
			OnlyToType:       lo.ToPtr(o.FromFilterOptions.ToType),
			OnlyRelationType: lo.ToPtr(req.RelationType),
		}, p, sort.TimeCursorType); err != nil {
			return resp, err
		}

		if p.LastToken != nil {
			resp.Token = *p.LastToken
		}
		resp.Total = total
		resp.Relations = lo.Map(relations, func(comment *relationmapper.Relation, _ int) *platform.Relation {
			return convertor.RelationMapperToRelation(comment)
		})
	case *platform.GetRelationsReq_ToFilterOptions:
		if relations, total, err = s.RelationMongoMapper.FindManyAndCount(ctx, &relationmapper.FilterOptions{
			OnlyFromType:     lo.ToPtr(o.ToFilterOptions.FromType),
			OnlyToId:         lo.ToPtr(o.ToFilterOptions.ToId),
			OnlyToType:       lo.ToPtr(o.ToFilterOptions.ToType),
			OnlyRelationType: lo.ToPtr(req.RelationType),
		}, p, sort.TimeCursorType); err != nil {
			return resp, err
		}

		if p.LastToken != nil {
			resp.Token = *p.LastToken
		}
		resp.Total = total
		resp.Relations = lo.Map(relations, func(comment *relationmapper.Relation, _ int) *platform.Relation {
			return convertor.RelationMapperToRelation(comment)
		})
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *RelationServiceImpl) DeleteRelation(ctx context.Context, req *platform.DeleteRelationReq) (resp *platform.DeleteRelationResp, err error) {
	resp = new(platform.DeleteRelationResp)

	if _, err = s.RelationMongoMapper.Delete(ctx, &relationmapper.FilterOptions{
		OnlyFromType:     lo.ToPtr(req.FromType),
		OnlyFromId:       lo.ToPtr(req.FromId),
		OnlyToType:       lo.ToPtr(req.ToType),
		OnlyToId:         lo.ToPtr(req.ToId),
		OnlyRelationType: lo.ToPtr(req.RelationType),
	}); err != nil {
		return resp, err
	}

	if s.Config.Neo4jConf.Enable {
		if err = s.RelationModel.DeleteEdge(ctx, &platform.Relation{
			FromType:     req.FromType,
			FromId:       req.FromId,
			ToType:       req.ToType,
			ToId:         req.ToId,
			RelationType: req.RelationType,
		}); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func (s *RelationServiceImpl) CreateRelation(ctx context.Context, req *platform.CreateRelationReq) (resp *platform.CreateRelationResp, err error) {
	resp = new(platform.CreateRelationResp)

	var res *platform.GetRelationResp
	if res, err = s.GetRelation(ctx, &platform.GetRelationReq{
		FromType:     req.FromType,
		FromId:       req.FromId,
		ToType:       req.ToType,
		ToId:         req.ToId,
		RelationType: req.RelationType,
	}); err != nil {
		return resp, err
	}
	fmt.Println(res.Ok)
	if !res.Ok {
		if _, err = s.RelationMongoMapper.Insert(ctx, &relationmapper.Relation{
			ID:           primitive.NilObjectID,
			FromType:     req.FromType,
			FromId:       req.FromId,
			ToType:       req.ToType,
			ToId:         req.ToId,
			RelationType: req.RelationType,
		}); err != nil {
			return resp, err
		}

		if s.Config.Neo4jConf.Enable {
			if err = s.RelationModel.CreateEdge(ctx, &platform.Relation{
				FromType:     req.FromType,
				FromId:       req.FromId,
				ToType:       req.ToType,
				ToId:         req.ToId,
				RelationType: req.RelationType,
			}); err != nil {
				return resp, err
			}
		}
		resp.Ok = true
	}
	return resp, nil
}

func (s *RelationServiceImpl) GetRelation(ctx context.Context, req *platform.GetRelationReq) (resp *platform.GetRelationResp, err error) {
	resp = new(platform.GetRelationResp)

	var relation *relationmapper.Relation
	if relation, err = s.RelationMongoMapper.FindOne(ctx, &relationmapper.FilterOptions{
		OnlyFromType:     lo.ToPtr(req.FromType),
		OnlyFromId:       lo.ToPtr(req.FromId),
		OnlyToType:       lo.ToPtr(req.ToType),
		OnlyToId:         lo.ToPtr(req.ToId),
		OnlyRelationType: lo.ToPtr(req.RelationType),
	}); err != nil && !errors.Is(err, consts.ErrNotFound) {
		return resp, err
	}
	fmt.Println(err, relation)
	if relation != nil {
		resp.Ok = true
	}

	return resp, nil
}
