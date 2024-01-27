package comment

import (
	"context"
	errorx "errors"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/pagination/mongop"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/zeromicro/go-zero/core/mr"
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

const CollectionName = "comment"

var prefixCommentCacheKey = "cache:comment:"
var _ IMongoMapper = (*MongoMapper)(nil)

type (
	IMongoMapper interface {
		Insert(ctx context.Context, data *Comment) (string, error)
		FindOne(ctx context.Context, fopts *FilterOptions) (*Comment, error)
		Update(ctx context.Context, data *Comment) (*mongo.UpdateResult, error)
		UpdateAfterCreateComment(ctx context.Context, data *Comment)
		Delete(ctx context.Context, id string) (int64, error)
		DeleteWithUserId(ctx context.Context, id, userId string) (int64, error)
		Count(ctx context.Context, filter *FilterOptions) (int64, error)
		FindMany(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*Comment, error)
		FindManyAndCount(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*Comment, int64, error)
		GetConn() *monc.Model
		StartClient() *mongo.Client
	}

	Comment struct {
		ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		UserId    string             `bson:"userId,omitempty" json:"userId,omitempty"`
		AtUserId  string             `bson:"atUserId,omitempty" json:"atUserId,omitempty"`
		SubjectId string             `bson:"subjectId,omitempty" json:"subjectId,omitempty"`
		RootId    string             `bson:"rootId,omitempty" json:"rootId,omitempty"`
		FatherId  string             `bson:"fatherId,omitempty" json:"fatherId,omitempty"`
		Content   string             `bson:"content,omitempty" json:"content,omitempty"`
		Meta      string             `bson:"meta,omitempty" json:"meta,omitempty"`
		Tags      []string           `bson:"tags,omitempty" json:"tags,omitempty"`
		Count     *int64             `bson:"count,omitempty" json:"count,omitempty"`
		State     int64              `bson:"state,omitempty" json:"state,omitempty"`
		Attrs     int64              `bson:"attrs,omitempty" json:"attrs,omitempty"`
		CreateAt  time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
		SortTime  int64              `bson:"sortTime,omitempty" json:"sortTime,omitempty"`
		HeatValue float64            `bson:"heatValue,omitempty" json:"heatValue,omitempty"`
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

func (m *MongoMapper) Insert(ctx context.Context, data *Comment) (string, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Insert", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.SortTime = data.CreateAt.UnixMilli()
	}

	key := prefixCommentCacheKey + data.ID.Hex()
	ID, err := m.conn.InsertOne(ctx, key, data)
	if err != nil {
		return "", err
	}
	return ID.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m *MongoMapper) FindOne(ctx context.Context, fopts *FilterOptions) (*Comment, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindOne", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	var data Comment
	if fopts.OnlyCommentId != nil {
		_, err := primitive.ObjectIDFromHex(*fopts.OnlyCommentId)
		if err != nil {
			return nil, consts.ErrInvalidId
		}
	}

	filter := makeMongoFilter(fopts)
	key := prefixCommentCacheKey + *fopts.OnlyCommentId
	if err := m.conn.FindOne(ctx, key, &data, filter); err != nil {
		if errorx.Is(err, monc.ErrNotFound) {
			return nil, consts.ErrNotFound
		} else {
			return nil, err
		}
	}
	return &data, nil
}

func (m *MongoMapper) Update(ctx context.Context, data *Comment) (*mongo.UpdateResult, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Update", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	key := prefixCommentCacheKey + data.ID.Hex()
	res, err := m.conn.UpdateOne(ctx, key, bson.M{consts.ID: data.ID, consts.UserId: data.UserId}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) UpdateAfterCreateComment(ctx context.Context, data *Comment) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.UpdateAfterCreateComment", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	update := bson.M{"$inc": bson.M{}}
	if data.Count != nil {
		update["$inc"].(bson.M)[consts.Count] = *data.Count
	}

	key := prefixCommentCacheKey + data.ID.Hex()
	_, _ = m.conn.UpdateOne(ctx, key, bson.M{consts.ID: data.ID}, update)
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

func (m *MongoMapper) DeleteWithUserId(ctx context.Context, id, userId string) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.DeleteWithUserId", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, consts.ErrInvalidId
	}
	key := prefixCommentCacheKey + id
	resp, err := m.conn.DeleteOne(ctx, key, bson.M{consts.ID: oid, consts.UserId: userId})
	return resp, err
}

func (m *MongoMapper) Count(ctx context.Context, fopts *FilterOptions) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Count", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	filter := makeMongoFilter(fopts)
	return m.conn.CountDocuments(ctx, filter)
}

func (m *MongoMapper) FindMany(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*Comment, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindMany", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	p := mongop.NewMongoPaginator(pagination.NewRawStore(sorter), popts)
	filter := makeMongoFilter(fopts)
	sort, err := p.MakeSortOptions(ctx, filter)
	if err != nil {
		return nil, err
	}

	var data []*Comment
	if err = m.conn.Find(ctx, &data, filter, &options.FindOptions{
		Sort:  sort,
		Limit: popts.Limit,
		Skip:  popts.Offset,
	}); err != nil {
		if errorx.Is(err, monc.ErrNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}

	// 如果是反向查询，反转数据
	if *popts.Backward {
		for i := 0; i < len(data)/2; i++ {
			data[i], data[len(data)-i-1] = data[len(data)-i-1], data[i]
		}
	}
	if len(data) > 0 {
		if err = p.StoreCursor(ctx, data[0], data[len(data)-1]); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (m *MongoMapper) FindManyAndCount(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*Comment, int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindManyAndCount", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	var data []*Comment
	var total int64
	var err, err1, err2 error
	err = mr.Finish(func() error {
		data, err1 = m.FindMany(ctx, fopts, popts, sorter)
		return err1
	}, func() error {
		total, err2 = m.Count(ctx, fopts)
		return err2
	})
	return data, total, err
}

func (m *MongoMapper) GetConn() *monc.Model {
	return m.conn
}

func (m *MongoMapper) StartClient() *mongo.Client {
	return m.conn.Database().Client()
}
