package provider

import (
	"github.com/CloudStriver/platform/biz/application/service"
	"github.com/CloudStriver/platform/biz/infrastructure/config"
	"github.com/CloudStriver/platform/biz/infrastructure/kq"
	commentModel "github.com/CloudStriver/platform/biz/infrastructure/mapper/comment"
	labelModel "github.com/CloudStriver/platform/biz/infrastructure/mapper/label"
	"github.com/CloudStriver/platform/biz/infrastructure/mapper/relation"
	subjectModel "github.com/CloudStriver/platform/biz/infrastructure/mapper/subject"
	"github.com/CloudStriver/platform/biz/infrastructure/stores/redis"
	"github.com/google/wire"
)

var AllProvider = wire.NewSet(
	ApplicationSet,
	InfrastructureSet,
)

var ApplicationSet = wire.NewSet(
	service.CommentSet,
	service.SubjectSet,
	service.LabelSet,
	service.RelationSet,
)

var InfrastructureSet = wire.NewSet(
	config.NewConfig,
	redis.NewRedis,
	kq.NewDeleteCommentRelationKq,
	MapperSet,
)

var MapperSet = wire.NewSet(
	commentModel.NewMongoMapper,
	subjectModel.NewMongoMapper,
	labelModel.NewMongoMapper,
	labelModel.NewEsMapper,
	relation.NewNeo4jMapper,
	relation.NewMongoMapper,
)
