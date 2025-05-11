package endpoints

import (
	"context"

	"github.com/google/uuid"
	allManager "github.com/yash3004/user_management_service"
)

type UsersEndpoint struct {
	UsersManager allManager.UserManager
}

func (ep *UsersEndpoint) ListUsers(ctx context.Context, request interface{}) (interface{}, error) {
	projectId := uuid.UUID{}
	userId := uuid.UUID{}
	return ep.UsersManager.GetUserByID(ctx, projectId, userId)
}
