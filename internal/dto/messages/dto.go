package messages_dto

type GetMessageQueryParams struct {
	ConvId         int64
	Limit          int
	LastReceivedId *int64
	UserId         int64
}
