package label

import (
	"context"
	errorx "errors"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

const CollectionName = "label"

var prefixCommentCacheKey = "cache:label:"
var _ IMongoMapper = (*MongoMapper)(nil)

type (
	IMongoMapper interface {
		Insert(ctx context.Context, data *Label) (string, error)
		FindOne(ctx context.Context, id string) (*Label, error)
		FindManyByIds(ctx context.Context, ids []string) ([]*Label, error)
		Update(ctx context.Context, data *Label) (*mongo.UpdateResult, error)
		Delete(ctx context.Context, id string) (int64, error)
		GetConn() *monc.Model
		StartClient() *mongo.Client
	}

	Label struct {
		ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		Value    string             `bson:"value,omitempty" json:"value,omitempty"`
		CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
		UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
		Score_   float64            `bson:"score_,omitempty" json:"score_,omitempty"`
	}

	MongoMapper struct {
		conn *monc.Model
	}
)

func NewMongoMapper(config *config.Config) IMongoMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, CollectionName, config.CacheConf)
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"value": 1, // 索引字段
		},
		Options: options.Index().SetUnique(true), // 唯一性索引
	}
	_, err := conn.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Error("labelModel Unique index created err[%v]\n", err)
	} else {
		log.Info("labelModel Unique index created successfully\n")
	}

	return &MongoMapper{
		conn: conn,
	}
}

func (m *MongoMapper) Insert(ctx context.Context, data *Label) (string, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Insert", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	key := prefixCommentCacheKey + data.ID.Hex()
	ID, err := m.conn.InsertOne(ctx, key, data)
	if err != nil {
		return "", err
	}
	return ID.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *MongoMapper) FindOne(ctx context.Context, id string) (*Label, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindOne", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidId
	}

	var data Label
	key := prefixCommentCacheKey + id
	err = m.conn.FindOne(ctx, key, &data, bson.M{consts.ID: oid})
	switch {
	case errorx.Is(err, monc.ErrNotFound):
		return nil, consts.ErrNotFound
	case err == nil:
		return &data, nil
	default:
		return nil, err
	}
}

func (m *MongoMapper) FindManyByIds(ctx context.Context, ids []string) ([]*Label, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindManyByIds", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	var data []*Label
	filter := bson.M{
		consts.ID: bson.M{
			"$in": lo.Map[string, primitive.ObjectID](ids, func(s string, _ int) primitive.ObjectID {
				oid, _ := primitive.ObjectIDFromHex(s)
				return oid
			}),
		},
	}
	if err := m.conn.Find(ctx, &data, filter); err != nil {
		if errorx.Is(err, monc.ErrNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (m *MongoMapper) Update(ctx context.Context, data *Label) (*mongo.UpdateResult, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Update", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	data.UpdateAt = time.Now()
	key := prefixCommentCacheKey + data.ID.Hex()
	res, err := m.conn.UpdateOne(ctx, key, bson.M{consts.ID: data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) Delete(ctx context.Context, id string) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Delete", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, consts.ErrInvalidId
	}
	key := prefixCommentCacheKey + id
	resp, err := m.conn.DeleteOne(ctx, key, bson.M{consts.ID: oid})
	return resp, err
}

func (m *MongoMapper) GetConn() *monc.Model {
	return m.conn
}

func (m *MongoMapper) StartClient() *mongo.Client {
	return m.conn.Database().Client()
}
