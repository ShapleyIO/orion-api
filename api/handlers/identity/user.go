package identity

import (
	"github.com/oapi-codegen/runtime/types"
)

type UserWithPassword struct {
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     types.Email `json:"email"`
	Password  string      `json:"password"`
}
