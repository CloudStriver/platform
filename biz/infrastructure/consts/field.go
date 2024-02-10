package consts

const (
	ID         = "_id"
	UserId     = "userId"
	AtUserId   = "atUserId"
	SubjectId  = "subjectId"
	RootId     = "rootId"
	FatherId   = "fatherId"
	Content    = "content"
	Meta       = "meta"
	Tags       = "tags"
	Count      = "count"
	ItemId     = "itemId"
	RootCount  = "rootCount"
	AllCount   = "allCount"
	State      = "state"
	Attrs      = "attrs"
	CreateAt   = "createAt"
	UpdateAt   = "updateAt"
	Value      = "value"
	Labels     = "labels"
	ObjectType = "objectType"
)

const (
	UnknownState         int64 = 0
	None                 int64 = 1 // 无
	Pinned               int64 = 2 // 置顶
	Highlighted          int64 = 3 // 精华
	PinnedAndHighlighted int64 = 4 //  置顶+精华
	Normal               int64 = 1
	Deleted              int64 = 2
)
