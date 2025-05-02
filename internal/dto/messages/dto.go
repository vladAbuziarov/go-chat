package messages_dto

import "chatapp/internal/entities/users"

type GetMessageQueryParams struct {
	ConvId         int64
	Limit          int
	LastReceivedId *int64
	UserId         users.UserId
}
