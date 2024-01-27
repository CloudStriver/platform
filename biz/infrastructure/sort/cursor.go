package sort

import (
	"go.mongodb.org/mongo-driver/bson"
	"math"
)

type (
	MongoCursor interface {
		MakeSortOptions(filter bson.M, backward bool) (bson.M, error)
	}

	TimeCursor struct {
		SortTime int64 `json:"sortTime"`
	}
)

var (
	TimeCursorType = (*TimeCursor)(nil)
)

func (s *TimeCursor) MakeSortOptions(filter bson.M, backward bool) (bson.M, error) {
	//构造lastId
	var sortTime int64
	if s == nil {
		if backward {
			sortTime = 0
		} else {
			sortTime = math.MaxInt64
		}
	} else {
		sortTime = s.SortTime
	}

	var sort bson.M
	if backward {
		filter["sortTime"] = bson.M{"$gt": sortTime}
		sort = bson.M{"sortTime": 1}
	} else {
		filter["sortTime"] = bson.M{"$lt": sortTime}
		sort = bson.M{"sortTime": -1}
	}
	return sort, nil
}
