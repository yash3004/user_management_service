package endpoints

import (
	allManager "github.com/yash3004/user_management_service"
)

type UsersEndpoint struct {
	UsersManager allManager.UserManager
}
