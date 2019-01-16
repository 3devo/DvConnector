package responses

import (
	"github.com/3devo/feconnector/models"
)

// ConfigCreationBody is the body that is needed to create a new Config through rest
// swagger:parameters CreateConfig UpdateConfig
type ConfigCreationBody struct {
	//in:body
	Data models.Config `json:"data"`
}
