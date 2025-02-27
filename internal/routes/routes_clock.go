package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
)

// Clock clock endpoint
func Clock(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"ShowTime", requestConfig.ShowTime,
			"TimeFormat", requestConfig.TimeFormat,
			"ShowDate", requestConfig.ShowDate,
			"DateFormat", requestConfig.DateFormat,
		)

		return Render(c, http.StatusOK, partials.Clock(requestConfig))
	}
}
