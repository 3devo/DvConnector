package responses

// LoginSuccess is successful loggingresponse
//
// This is used for returning the new JWT token to to the user
//
// swagger:model LoginSuccessResponse
type LoginSuccess struct {
	Data struct {
		Token string `json:"token"`
	}
}

// LoginParameters is the body that is needed to login the user
// swagger:parameters Login
type LoginParameters struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// AuthEnabledResponse is the response body for the authEnabled endpoint
// swagger:model AuthEnabledResponse
type AuthEnabledResponse struct {
	Enabled bool `json:"enabled"`
}
