package consts

import (
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidId             = status.Error(10101, "objectId无效")
	ErrPaginatorTokenExpired = status.Error(10102, "分页token已过期")
	ErrNotFound              = status.Error(10105, "数据不存在")
	ErrDataBase              = status.Error(10007, "数据库异常")
	ErrEsMapper              = status.Error(10008, "Es异常")
	ErrIllegalOperation      = status.Error(10009, "非法操作")
)
