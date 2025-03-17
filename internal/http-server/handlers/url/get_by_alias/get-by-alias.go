package getbyalias

import (
	"errors"
	"log/slog"
	"net/http"

	resp "ex.com/internal/lib/api/response"
	"ex.com/internal/lib/loggeer/sl"
	"ex.com/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	resp.Response
	Url string `json:"url"`
}

type GetUrlByAliaser interface {
	GetUrlByAlias(alias string) (string, error)
}

func New(log *slog.Logger, getter GetUrlByAliaser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.get_by_alias"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		alias := req.Alias
		res, err := getter.GetUrlByAlias(alias)
		if errors.Is(err, storage.ERR_NO_ALIAS_FOUND) {
			log.Error("there is no such alias in db")
			render.JSON(w, r, Response{resp.Error("there is no such alias in db"), ""})
		}

		if err != nil {
			log.Error("failed to get url by alias", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to get url by alias"))
			return
		}

		log.Info("Url by alias was got successfuly!", slog.Any("url", res))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Url:      res,
		})

	}
}
