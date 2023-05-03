package requests

import (
	"fmt"
	"github.com/go-chi/render"
	"net/http"
)

type PutValueRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewPutValueRequest(r *http.Request) (*PutValueRequest, error) {
	res := &PutValueRequest{}
	err := render.Bind(r, res)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	return res, nil
}

func (*PutValueRequest) Bind(r *http.Request) error {
	return nil
}
