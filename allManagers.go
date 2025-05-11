package usermanagementservice

import (
	"context"

	"github.com/google/uuid"
)

type User struct{
	UserId uuid.UUID
	UserName string
	Email string

}

type UserManager interface {
	GetUserByID(ctx context.Context,ProjectId uuid.UUID,UserID uuid.UUID)(User,error)
}

type ProjectManager interface {
}

type PoliciesManager interface {
}

type RolesManger interface {
}
