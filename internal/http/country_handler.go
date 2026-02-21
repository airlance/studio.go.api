package http

import (
	"net/http"

	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CountryHandler struct {
	countryUC usecase.CountryUseCase
}

func NewCountryHandler(countryUC usecase.CountryUseCase) *CountryHandler {
	return &CountryHandler{countryUC: countryUC}
}

func (h *CountryHandler) List(c *gin.Context) {
	countries, err := h.countryUC.ListAll(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list countries")
		httpx.RespondError(c, http.StatusInternalServerError, "internal_error", "Failed to list countries")
		return
	}

	resp := make([]*CountryResponse, 0, len(countries))
	for _, country := range countries {
		resp = append(resp, toCountryResponse(country))
	}

	httpx.RespondOK(c, gin.H{"countries": resp})
}
