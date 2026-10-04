package orders

import (
	"encoding/json"
	"net/http"
)

type handler struct {
}

func NewHandler() *handler {
	return &handler{}
}

func (this *handler) postOrder(w http.ResponseWriter, r *http.Request) {
	order := &Model{}
	json.NewDecoder(r.Body).Decode(order)

}
