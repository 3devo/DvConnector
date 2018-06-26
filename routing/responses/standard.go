package responses

import (
	"encoding/json"
	"net/http"
)

// A  error or success response model
// This is used to indicate errors or success messages
//
// swagger:response ResourceStatusResponse
type ResourceStatusResponse struct {
	// in: body
	Body struct {
		Code     int32  `json:"code"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
		Error    string `json:"error"`
	} `json:"body"`
}

//swagger:parameters GetLogFile UpdateLogFile DeleteLogFile GetChart UpdateChart DeleteChart GetSheet UpdateSheet DeleteSheet GetWorkspace UpdateWorkspace DeleteWorkspace
type UidPathParam struct {
	// in: path
	UUID string `json:"uuid"`
}

func WriteResourceStatusResponse(code int32, resource string, action string, err string, w http.ResponseWriter) {
	var response ResourceStatusResponse
	response.Body.Code = code
	response.Body.Resource = resource
	response.Body.Action = action
	response.Body.Error = err
	json.NewEncoder(w).Encode(response.Body)
}
