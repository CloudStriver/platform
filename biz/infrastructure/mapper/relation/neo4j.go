package relation

import (
	"context"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	genrelation "github.com/CloudStriver/service-idl-gen-go/kitex_gen/platform/relation"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/zeromicro/go-zero/core/mr"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

const EdgeName = "relation"

var _ RelationNeo4jMapper = (*Neo4jMapper)(nil)

type (
	RelationNeo4jMapper interface {
		CreateEdge(ctx context.Context, relation *genrelation.Relation) error
		MatchEdge(ctx context.Context, relation *genrelation.Relation) (bool, error)
		DeleteEdge(ctx context.Context, relation *genrelation.Relation) error
		MatchFromEdges(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error)
		MatchToEdges(ctx context.Context, toType int64, toId string, fromType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error)
		MatchFromEdgesCount(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64) (int64, error)
		MatchToEdgesCount(ctx context.Context, toType int64, toId string, fromType int64, relationType int64) (int64, error)
		MatchFromEdgesAndCount(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, int64, error)
		MatchToEdgesAndCount(ctx context.Context, toType int64, toId string, fromType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, int64, error)
		GetRelationPaths(ctx context.Context, fromType int64, fromId string, type1 int64, type2 int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error)
		DeleteNode(ctx context.Context, fromId string, fromType int64) error
	}
	Neo4jMapper struct {
		conn     neo4j.DriverWithContext
		DataBase string
	}
)

func (n *Neo4jMapper) DeleteNode(ctx context.Context, fromId string, fromType int64) error {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.DeleteNode", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	_, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MATCH (n1:node{name: $FromId, type: $FromType}) DETACH DELETE n1",
		map[string]any{
			"FromId":   fromId,
			"FromType": fromType,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase),
	)
	if err != nil {
		log.CtxError(ctx, "创建关系异常[%v]\n", err)
		return err
	}
	return nil
}

func (n *Neo4jMapper) GetRelationPaths(ctx context.Context, fromType int64, fromId string, type1 int64, type2 int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error) {
	options.EnsureSafe()
	result, err := neo4j.ExecuteQuery(ctx, n.conn, "MATCH path=(node1:node {name: $FromId, type: $FromType})-[r1:edge {type: $Type1}]->(node2:node)-[r2:edge {type: $Type2}]->(node3:node) RETURN node2,node3,r2 ORDER BY r2.createTime DESC SKIP $Offset LIMIT $Limit",
		map[string]any{
			"Offset":   *options.Offset,
			"Limit":    *options.Limit,
			"FromId":   fromId,
			"FromType": fromType,
			"Type1":    type1,
			"Type2":    type2,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return nil, err
	}
	relation := make([]*genrelation.Relation, 0, len(result.Records))
	for i := range result.Records {
		relationNode, _, err := neo4j.GetRecordValue[neo4j.Relationship](result.Records[i], "r2")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		fromNode, _, err := neo4j.GetRecordValue[neo4j.Node](result.Records[i], "node2")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}
		fromId, err := neo4j.GetProperty[string](fromNode, "name")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}
		fromType, err := neo4j.GetProperty[int64](fromNode, "type")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		toNode, _, err := neo4j.GetRecordValue[neo4j.Node](result.Records[i], "node3")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}
		toId, err := neo4j.GetProperty[string](toNode, "name")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}
		toType, err := neo4j.GetProperty[int64](toNode, "type")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		createTime, err := neo4j.GetProperty[int64](relationNode, "createTime")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		relation = append(relation, &genrelation.Relation{
			FromType:     fromType,
			FromId:       fromId,
			ToType:       toType,
			ToId:         toId,
			RelationType: type2,
			CreateTime:   createTime,
		})
	}
	return relation, nil
}

func (n *Neo4jMapper) MatchFromEdgesAndCount(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, int64, error) {
	options.EnsureSafe()
	relation := make([]*genrelation.Relation, 0, *options.Limit)
	var (
		count           int64
		err, err1, err2 error
	)

	if err = mr.Finish(func() error {
		relation, err1 = n.MatchFromEdges(ctx, fromType, fromId, toType, relationType, options)
		if err1 != nil {
			return err1
		}
		return nil
	}, func() error {
		count, err2 = n.MatchFromEdgesCount(ctx, fromType, fromId, toType, relationType)
		if err2 != nil {
			return err2
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	return relation, count, nil
}

func (n *Neo4jMapper) MatchToEdgesAndCount(ctx context.Context, toType int64, toId string, fromType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, int64, error) {
	options.EnsureSafe()
	relation := make([]*genrelation.Relation, 0, *options.Limit)
	var count int64
	var err error
	if err = mr.Finish(func() error {
		relation, err = n.MatchToEdges(ctx, toType, toId, fromType, relationType, options)
		if err != nil {
			return err
		}
		return nil
	}, func() error {
		count, err = n.MatchToEdgesCount(ctx, toType, toId, fromType, relationType)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	return relation, count, nil
}

func (n *Neo4jMapper) MatchFromEdgesCount(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.MatchFromEdgesCount", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	result, err := neo4j.ExecuteQuery(ctx, n.conn, "MATCH (n1:node{name: $FromId, type: $FromType})-[r:edge{type: $RelationType}]->(n2:node{type: $ToType}) RETURN COUNT(r)",
		map[string]any{
			"FromId":       fromId,
			"FromType":     fromType,
			"ToType":       toType,
			"RelationType": relationType,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return 0, err
	}
	count, _, err := neo4j.GetRecordValue[int64](result.Records[0], "COUNT(r)")
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (n *Neo4jMapper) MatchToEdgesCount(ctx context.Context, toType int64, toId string, fromType int64, relationType int64) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.MatchToEdgesCount", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	result, err := neo4j.ExecuteQuery(ctx, n.conn, "MATCH (n1:node{type: $FromType})-[r:edge{type: $RelationType}]->(n2:node{name: $ToId, type: $ToType}) RETURN COUNT(r)",
		map[string]any{
			"ToId":         toId,
			"ToType":       toType,
			"FromType":     fromType,
			"RelationType": relationType,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return 0, err
	}
	count, _, err := neo4j.GetRecordValue[int64](result.Records[0], "COUNT(r)")
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (n *Neo4jMapper) MatchFromEdges(ctx context.Context, fromType int64, fromId string, toType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.MatchFromEdges", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	options.EnsureSafe()
	result, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MATCH (n1:node{name: $FromId, type: $FromType})-[r:edge{type: $RelationType}]->(n2:node{type:$ToType}) RETURN n2,r ORDER BY r.createTime DESC SKIP $Offset LIMIT $Limit",
		map[string]any{
			"Offset":       *options.Offset,
			"Limit":        *options.Limit,
			"FromType":     fromType,
			"FromId":       fromId,
			"ToType":       toType,
			"RelationType": relationType,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return nil, err
	}

	relation := make([]*genrelation.Relation, 0, len(result.Records))
	for i := range result.Records {
		relationNode, _, err := neo4j.GetRecordValue[neo4j.Relationship](result.Records[i], "r")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		toNode, _, err := neo4j.GetRecordValue[neo4j.Node](result.Records[i], "n2")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		toId, err := neo4j.GetProperty[string](toNode, "name")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		createTime, err := neo4j.GetProperty[int64](relationNode, "createTime")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		relation = append(relation, &genrelation.Relation{
			ToType:       toType,
			ToId:         toId,
			FromType:     fromType,
			FromId:       fromId,
			RelationType: relationType,
			CreateTime:   createTime,
		})
	}
	return relation, nil
}

func (n *Neo4jMapper) MatchToEdges(ctx context.Context, toType int64, toId string, fromType int64, relationType int64, options *pagination.PaginationOptions) ([]*genrelation.Relation, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.MatchToEdges", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	options.EnsureSafe()
	result, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MATCH (n1:node{type:$FromType})-[r:edge{type:$RelationType}]->(n2:node{name:$ToId,type:$ToType}) RETURN n1,r ORDER BY r.createTime DESC SKIP $Offset LIMIT $Limit",
		map[string]any{
			"Offset":       *options.Offset,
			"Limit":        *options.Limit,
			"ToType":       toType,
			"ToId":         toId,
			"FromType":     fromType,
			"RelationType": relationType,
		},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return nil, err
	}
	relation := make([]*genrelation.Relation, 0, len(result.Records))
	for i := range result.Records {
		relationNode, _, err := neo4j.GetRecordValue[neo4j.Relationship](result.Records[i], "r")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		fromNode, _, err := neo4j.GetRecordValue[neo4j.Node](result.Records[i], "n1")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}
		fromId, err := neo4j.GetProperty[string](fromNode, "name")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		createTime, err := neo4j.GetProperty[int64](relationNode, "createTime")
		if err != nil {
			log.CtxError(ctx, "GetRecordValue异常[%v]\n", err)
			return nil, err
		}

		relation = append(relation, &genrelation.Relation{
			FromType:     fromType,
			FromId:       fromId,
			ToType:       toType,
			ToId:         toId,
			RelationType: relationType,
			CreateTime:   createTime,
		})
	}
	return relation, nil
}

func (n *Neo4jMapper) CreateEdge(ctx context.Context, relation *genrelation.Relation) error {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.CreateEdge", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	if _, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MERGE (n1:node{name: $FromId, type: $FromType}) MERGE (n2:node{name: $ToId, type: $ToType}) CREATE (n1)-[r:edge{type: $RelationType, createTime:$CreateTime}]->(n2)",
		map[string]any{
			"FromType":     relation.FromType,
			"FromId":       relation.FromId,
			"ToId":         relation.ToId,
			"ToType":       relation.ToType,
			"RelationType": relation.RelationType,
			"CreateTime":   time.Now().UnixMilli(),
		}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase)); err != nil {
		log.CtxError(ctx, "创建关系异常[%v]\n", err)
		return err
	}
	return nil
}

func (n *Neo4jMapper) MatchEdge(ctx context.Context, relation *genrelation.Relation) (bool, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.MatchEdge", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	result, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MATCH (n1:node{name: $FromId, type: $FromType})-[r:edge{type: $RelationType}]->(n2:node{name: $ToId, type: $ToType}) return r",
		map[string]any{
			"FromId":       relation.FromId,
			"FromType":     relation.FromType,
			"ToId":         relation.ToId,
			"ToType":       relation.ToType,
			"RelationType": relation.RelationType,
		}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase))
	if err != nil {
		log.CtxError(ctx, "查询关系异常[%v]\n", err)
		return false, err
	}

	return len(result.Records) != 0, nil
}

func (n *Neo4jMapper) DeleteEdge(ctx context.Context, relation *genrelation.Relation) error {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "neo4j.DeleteEdge", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	if _, err := neo4j.ExecuteQuery(ctx, n.conn,
		"MATCH (n1:node{name: $FromId, type: $FromType})-[r:edge{type: $RelationType}]->(n2:node{name: $ToId, type: $ToType}) DELETE r",
		map[string]any{
			"FromId":       relation.FromId,
			"FromType":     relation.FromType,
			"ToId":         relation.ToId,
			"ToType":       relation.ToType,
			"RelationType": relation.RelationType,
		}, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(n.DataBase)); err != nil {
		log.CtxError(ctx, "创建关系异常[%v]\n", err)
		return err
	}
	return nil
}

func NewNeo4jMapper(config *config.Config) RelationNeo4jMapper {
	conn, err := neo4j.NewDriverWithContext(config.Neo4jConf.Url, neo4j.BasicAuth(config.Neo4jConf.Username, config.Neo4jConf.Password, ""))
	if err != nil {
		panic(err)
	}
	return &Neo4jMapper{
		conn:     conn,
		DataBase: config.Neo4jConf.DataBase,
	}
}
