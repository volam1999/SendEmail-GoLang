package email

import (
	"net/http"

	"github.com/volam1999/gomail/internal/pkg/http/router"
)

func (h *Handler) Routes() []router.Route {
	return []router.Route{
		{
			Path:    "/api/v1/emails",
			Method:  http.MethodGet,
			Handler: h.FindAll,
		},
		{
			Path:    "/api/v1/emails/{id}",
			Method:  http.MethodGet,
			Handler: h.FindByEmailId,
		},
		{
			Path:    "/api/v1/emails/send",
			Method:  http.MethodPost,
			Handler: h.SendEmail,
		},
	}
}
