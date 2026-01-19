package handlers

import "net/http"

type DialogHandler interface {
}

type dialogHandler struct {
}

func (handler *dialogHandler) InitDialogHanler(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
}
