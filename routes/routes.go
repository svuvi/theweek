package routes

import (
	"database/sql"
	"embed"
	"net/http"

	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
	"github.com/svuvi/theweek/repositories"
)

type BaseHandler struct {
	articleRepo models.ArticleRepository
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		articleRepo: repositories.NewArticleRepo(db),
	}
}

//go:embed static
var static embed.FS

func (h *BaseHandler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", h.indexHandler)
	mux.HandleFunc("GET /{slug}", h.articleHandler)

	mux.Handle("GET /static/", http.FileServerFS(static))

	return mux
}

func (h *BaseHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := h.articleRepo.GetAll()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	layouts.Index(articles).Render(r.Context(), w)
}

func (h *BaseHandler) articleHandler(w http.ResponseWriter, r *http.Request) {

}
