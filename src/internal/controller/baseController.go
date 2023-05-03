package controller

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"webinar2/src/internal/requests"
)

type log interface {
	Info(args ...interface{})
}

type Storage interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type BaseController struct {
	logger  log
	storage Storage
}

func NewBaseController(logger log, storage Storage) *BaseController {
	return &BaseController{
		logger:  logger,
		storage: storage,
	}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", c.handleMain)
	r.Get("/{name}", c.handleName)
	r.Put("/", c.handleStore)

	return r
}

func (c *BaseController) handleMain(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("main page")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/html")
	w.Write([]byte("<h1>hello world</h1>"))
}

func (c *BaseController) handleName(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("get var")
	name := chi.URLParam(r, "name")
	value, err := c.storage.Get(name)
	if err != nil {
		c.logger.Info(fmt.Errorf("error when get value %s: %w", name, err))
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<h1>not found</h1>"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("<h1>error</h1>"))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	res := fmt.Sprintf("<h1>%s: %s</h1>", name, value)
	w.Write([]byte(res))
}

func (c *BaseController) handleStore(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("set var")
	val, err := requests.NewPutValueRequest(r)
	if err != nil {
		c.logger.Info(fmt.Errorf("request parse error: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<h1>error</h1>"))
		return
	}

	err = c.storage.Set(val.Name, val.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("<h1>error</h1>"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
