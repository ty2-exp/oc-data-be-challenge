package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"oc-data-be-challenge/internal/usecase"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-chi/render"
)

type ChiServer struct {
	dataPointUseCase *usecase.DataPointUseCase
}

func NewChiServer(dataPointUseCase *usecase.DataPointUseCase) *ChiServer {
	return &ChiServer{dataPointUseCase: dataPointUseCase}
}

func (chiServer ChiServer) DataPointQuery(w http.ResponseWriter, r *http.Request, params DataPointQueryParams) {
	start, err := chiServer.parseTime(params.Start)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, Error{
			Message: fmt.Errorf("failed to parse start time: %v", err).Error(),
		})
		return
	}

	until, err := chiServer.parseTime(params.Until)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, Error{
			Message: fmt.Errorf("failed to parse until time: %v", err).Error(),
		})
		return
	}

	resultIter, err := chiServer.dataPointUseCase.Query(r.Context(), start, until)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, Error{
			Message: fmt.Errorf("failed to query datapoints: %v", err).Error(),
		})
		return
	}

	// Set response header for JSON content
	w.Header().Set("Content-Type", "application/json")

	// Stream the JSON array directly to the response writer
	resBodyEncoder := sonic.Config{NoEncoderNewline: true}.Froze().NewEncoder(w)
	_, err = w.Write([]byte("["))
	if err != nil {
		slog.ErrorContext(r.Context(), "Error writing response start", "error", err)
		return
	}

	i := 0
	for resultIter.Next() {
		dp, err := resultIter.Value()
		if err != nil {
			slog.ErrorContext(r.Context(), "Error retrieving item", "iter_counter", i, "error", err)
			continue
		}

		// Add comma before all items except the first
		if i > 0 {
			_, err := w.Write([]byte(","))
			if err != nil {
				slog.ErrorContext(r.Context(), "Error writing response comma", "error", err)
				return
			}
		}

		// Encode the item directly to the response writer
		if err := resBodyEncoder.Encode(DataPointModel{
			Time:  dp.Time.Format(time.RFC3339),
			Value: dp.Value,
		}); err != nil {
			slog.ErrorContext(r.Context(), "Error encoding response item", "error", err)
			return
		}
		i++
	}
	_, err = w.Write([]byte("]"))
	if err != nil {
		slog.ErrorContext(r.Context(), "Error writing response end", "error", err)
		return
	}
}

func (chiServer ChiServer) parseTime(timeStr *string) (*time.Time, error) {
	if timeStr == nil {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *timeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w", err)
	}
	return &t, nil
}
