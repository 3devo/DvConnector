package responses

// A  error or success response model
// This is used to indicate errors or success messages
//
// swagger:response StatusResponse
type StatusResponse struct {
	// in: body
	Body struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"body"`
}

//swagger:parameters GetLogFile UpdateLogFile DeleteLogFile GetChart UpdateChart DeleteChart GetSheet UpdateSheet DeleteSheet GetWorkspace UpdateWorkspace DeleteWorkspace
type UidPathParam struct {
	// in: path
	UUID string `json:"uuid"`
}
