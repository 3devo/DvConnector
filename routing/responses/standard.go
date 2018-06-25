package responses

import (
	"encoding/json"
	"net/http"
)

// A  error or success response model
// This is used to indicate errors or success messages
//
// {body:{code:200, message:"log has been added successfully"}}
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

func WriteStatusResponse(code int32, message string, w http.ResponseWriter) {
	var response StatusResponse
	response.Body.Code = code
	response.Body.Message = message
	w.WriteHeader(int(code))
	json.NewEncoder(w).Encode(response)
}
