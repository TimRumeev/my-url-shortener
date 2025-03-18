package deleteurl

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
	Item resp.Url `json:"item"`
}

type Deleter interface {
	DeleteUrlByAlias(alias string) (resp.Url, error)
}

func New(log *slog.Logger, deleter Deleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.ErrorWithCode(r, 500, "failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ErrorWithCode(r, 400, "invalid request"))

			return
		}

		alias := req.Alias
		res, err := deleter.DeleteUrlByAlias(alias)
		if errors.Is(err, storage.ERR_URL_NOT_FOUND) {
			log.Error("there is no such alias in db", sl.Err(err))
			render.JSON(w, r, resp.ErrorWithCode(r, 404, "there is no such alias in db"))
			return
		}
		if err != nil {
			log.Error("failed to delete url by alias", sl.Err(err))

			render.JSON(w, r, resp.ErrorWithCode(r, 500, "failed to delete url by alias"))

			return
		}

		log.Info("Url was successfuly deleted!", slog.Any("Item", res))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Item:     res,
		})

	}
}
