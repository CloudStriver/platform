package subject

import (
	"context"
	errorx "errors"
	"github.com/CloudStriver/platform/biz/infrastructure/config"
	"github.com/CloudStriver/platform/biz/infrastructure/consts"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

const CollectionName = "subject"

var prefixSubjectCacheKey = "cache:subject:"
var _ IMongoMapper = (*MongoMapper)(nil)

type (
	IMongoMapper interface {
		Insert(ctx context.Context, data *Subject) (string, error)
		FindOne(ctx context.Context, id string) (*Subject, error)
		Update(ctx context.Context, data *Subject) (*mongo.UpdateResult, error)
		UpdateCount(ctx context.Context, id string, allCount, rootCount int64)
		Delete(ctx context.Context, id string) (int64, error)
		GetConn() *monc.Model
		StartClient() *mongo.Client
	}

	Subject struct {
		ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		UserId       string             `bson:"userId,omitempty" json:"userId,omitempty"`
		TopCommentId *string            `bson:"topCommentId,omitempty" json:"topCommentId,omitempty"`
		RootCount    *int64             `bson:"rootCount,omitempty" json:"rootCount,omitempty"`
		AllCount     *int64             `bson:"allCount,omitempty" json:"allCount,omitempty"`
		State        int64              `bson:"state,omitempty" json:"state,omitempty"`
		Attrs        int64              `bson:"attrs,omitempty" json:"attrs,omitempty"`
		CreateAt     time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
		UpdateAt     time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	}

	MongoMapper struct {
		conn *monc.Model
	}
)

func NewMongoMapper(config *config.Config) IMongoMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, CollectionName, config.CacheConf)
	return &MongoMapper{
		conn: conn,
	}
}

func (m *MongoMapper) Insert(ctx context.Context, data *Subject) (string, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Insert", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}
	data.CreateAt = time.Now()
	data.UpdateAt = time.Now()
	key := prefixSubjectCacheKey + data.ID.Hex()
	ID, err := m.conn.InsertOne(ctx, key, data)
	if err != nil {
		return "", err
	}
	return ID.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *MongoMapper) FindOne(ctx context.Context, id string) (*Subject, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindOne", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidId
	}
	var data Subject
	key := prefixSubjectCacheKey + id
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

func (m *MongoMapper) Update(ctx context.Context, data *Subject) (*mongo.UpdateResult, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Update", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()
	data.UpdateAt = time.Now()
	key := prefixSubjectCacheKey + data.ID.Hex()
	res, err := m.conn.UpdateOne(ctx, key, bson.M{consts.ID: data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) UpdateCount(ctx context.Context, id string, allCount, rootCount int64) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.UpdateAfterCreateComment", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, _ := primitive.ObjectIDFromHex(id)
	key := prefixSubjectCacheKey + id
	_, _ = m.conn.UpdateOne(ctx, key, bson.M{consts.ID: oid}, bson.M{"$inc": bson.M{consts.RootCount: rootCount, consts.AllCount: allCount}, "$set": bson.M{consts.UpdateAt: time.Now()}})
}

func (m *MongoMapper) Delete(ctx context.Context, id string) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Delete", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, consts.ErrInvalidId
	}
	key := prefixSubjectCacheKey + id
	resp, err := m.conn.DeleteOne(ctx, key, bson.M{consts.ID: oid})
	return resp, err
}

func (m *MongoMapper) GetConn() *monc.Model {
	return m.conn
}

func (m *MongoMapper) StartClient() *mongo.Client {
	return m.conn.Database().Client()
}
