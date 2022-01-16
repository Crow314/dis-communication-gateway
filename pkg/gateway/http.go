package gateway

import (
	"encoding/hex"
	"net/http"
	"strconv"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	srv, err := strconv.Atoi(r.PostFormValue("service"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := hex.DecodeString(r.PostFormValue("data"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(data) > 28 { // max length is (32-4) Byte = 28 Byte
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := transmitRequest{service: byte(srv), data: data, notificator: make(chan struct{})}
	transmitReqChan <- req

	<-req.notificator // 処理待機

	if req.errorResponse != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
