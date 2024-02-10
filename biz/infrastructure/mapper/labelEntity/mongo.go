package labelEntity

import (
	"context"
	errorx "errors"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/pagination/mongop"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/samber/lo"
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

const CollectionName = "labelEntity"

var prefixCommentCacheKey = "cache:labelEntity:"
var _ IMongoMapper = (*MongoMapper)(nil)

type (
	IMongoMapper interface {
		Insert(ctx context.Context, data *LabelEntity) error
		InsertMany(ctx context.Context, data []*LabelEntity) error
		FindOne(ctx context.Context, id string) (*LabelEntity, error)
		Count(ctx context.Context, filter *FilterOptions) (int64, error)
		FindMany(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*LabelEntity, error)
		FindManyAndCount(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*LabelEntity, int64, error)
		Update(ctx context.Context, data *LabelEntity) (*mongo.UpdateResult, error)
		Delete(ctx context.Context, id string) (int64, error)
		GetConn() *monc.Model
		StartClient() *mongo.Client
	}

	LabelEntity struct {
		ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		ObjectType int64              `bson:"objectType,omitempty" json:"objectType,omitempty"`
		Labels     []string           `bson:"labels,omitempty" json:"labels,omitempty"`
		CreateAt   time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
		UpdateAt   time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
		Score_     float64            `bson:"score_,omitempty" json:"score_,omitempty"`
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

func (m *MongoMapper) Insert(ctx context.Context, data *LabelEntity) error {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Insert", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	data.CreateAt = time.Now()
	data.UpdateAt = time.Now()
	key := prefixCommentCacheKey + data.ID.Hex()
	_, err := m.conn.InsertOne(ctx, key, data)
	return err
}

func (m *MongoMapper) InsertMany(ctx context.Context, data []*LabelEntity) error {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.InsertMany", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	for i := range data {
		data[i].CreateAt = time.Now()
		data[i].UpdateAt = time.Now()
	}

	dataAny := lo.Map(data, func(item *LabelEntity, _ int) any { return item })
	_, err := m.conn.InsertMany(ctx, dataAny)
	return err
}

func (m *MongoMapper) FindOne(ctx context.Context, id string) (*LabelEntity, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindOne", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidId
	}

	var data LabelEntity
	key := prefixCommentCacheKey + id
	if err = m.conn.FindOne(ctx, key, &data, bson.M{consts.ID: oid}); err != nil {
		if errorx.Is(err, monc.ErrNotFound) {
			return nil, consts.ErrNotFound
		} else {
			return nil, err
		}
	}
	return &data, nil
}

func (m *MongoMapper) Update(ctx context.Context, data *LabelEntity) (*mongo.UpdateResult, error) {
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

func (m *MongoMapper) Count(ctx context.Context, fopts *FilterOptions) (int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.Count", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	filter := makeMongoFilter(fopts)
	return m.conn.CountDocuments(ctx, filter)
}

func (m *MongoMapper) FindMany(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*LabelEntity, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindMany", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	p := mongop.NewMongoPaginator(pagination.NewRawStore(sorter), popts)
	filter := makeMongoFilter(fopts)
	sort, err := p.MakeSortOptions(ctx, filter)
	if err != nil {
		return nil, err
	}

	var data []*LabelEntity
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

func (m *MongoMapper) FindManyAndCount(ctx context.Context, fopts *FilterOptions, popts *pagination.PaginationOptions, sorter mongop.MongoCursor) ([]*LabelEntity, int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "mongo.FindManyAndCount", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	var data []*LabelEntity
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
