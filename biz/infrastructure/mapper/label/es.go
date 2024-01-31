package label

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/CloudStriver/go-pkg/utils/pagination"
	"github.com/CloudStriver/go-pkg/utils/pagination/esp"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/CloudStriver/platform-comment/biz/infrastructure/consts"
	"github.com/bytedance/sonic"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
	"time"
)

type (
	IEsMapper interface {
		Search(ctx context.Context, query []types.Query, popts *pagination.PaginationOptions, sorter esp.EsCursor) ([]*Label, int64, error)
	}

	EsMapper struct {
		es        *elasticsearch.TypedClient
		IndexName string
	}
)

func (e *EsMapper) Search(ctx context.Context, query []types.Query, popts *pagination.PaginationOptions, sorter esp.EsCursor) ([]*Label, int64, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	_, span := tracer.Start(ctx, "elasticsearch.Search", oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	defer span.End()

	p := esp.NewEsPaginator(pagination.NewRawStore(sorter), popts)
	s, sa, err := p.MakeSortOptions(ctx)
	if err != nil {
		log.CtxError(ctx, "创建索引异常[%v]\n", err)
		return nil, 0, err
	}
	res, err := e.es.Search().Index(e.IndexName).Request(&search.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: query,
			},
		},
		Sort:        s,
		SearchAfter: sa,
		Size:        lo.ToPtr(int(*popts.Limit)),
	}).Do(ctx)
	if err != nil {
		logx.Errorf("es查询异常[%v]\n", err)
		return nil, 0, err
	}

	total := res.Hits.Total.Value
	labels := make([]*Label, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		label := &Label{}
		source := make(map[string]any)
		err = sonic.Unmarshal(hit.Source_, &source)
		if err != nil {
			return nil, 0, err
		}
		if source[consts.CreateAt], err = time.Parse("2006-01-02T15:04:05Z07:00", source[consts.CreateAt].(string)); err != nil {
			return nil, 0, err
		}
		if source[consts.UpdateAt], err = time.Parse("2006-01-02T15:04:05Z07:00", source[consts.UpdateAt].(string)); err != nil {
			return nil, 0, err
		}
		err = mapstructure.Decode(source, label)
		if err != nil {
			return nil, 0, err
		}

		oid := hit.Id_
		label.ID, err = primitive.ObjectIDFromHex(oid)
		if err != nil {
			return nil, 0, consts.ErrInvalidId
		}
		label.Score_ = float64(hit.Score_)
		labels = append(labels, label)
	}

	if *popts.Backward {
		labels = lo.Reverse(labels)
	}

	// 更新游标
	if len(labels) > 0 {
		err = p.StoreCursor(ctx, labels[0], labels[len(labels)-1])
		if err != nil {
			return nil, 0, err
		}
	}
	return labels, total, nil
}

func NewEsMapper(config *config.Config) IEsMapper {
	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Username:  config.Elasticsearch.Username,
		Password:  config.Elasticsearch.Password,
		Addresses: config.Elasticsearch.Addresses,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		logx.Errorf("elasticsearch连接异常[%v]\n", err)
	}
	return &EsMapper{
		es:        esClient,
		IndexName: fmt.Sprintf("%s.%s", config.Mongo.DB, CollectionName),
	}
}
