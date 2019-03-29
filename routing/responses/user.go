package responses

import "github.com/3devo/dvconnector/models"

// UserCreationBody is the body needed to create a user through rest
// swagger:parameters CreateUser UpdateUser
type UserCreationBody struct {
	// in:body
	Data models.User `json:"data"`
}
