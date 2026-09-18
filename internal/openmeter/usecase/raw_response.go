package usecase

import (
	"context"
	"time"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
	"github.com/duborgis/llm-mock-api/internal/openmeter/ports"
)

// RawResponseIngest implements ports.RawResponseIngestUseCase. Pure business logic: no HTTP,
// no MongoDB driver.
type RawResponseIngest struct {
	Responses ports.RawResponseRepository
}

func NewRawResponseIngest(responses ports.RawResponseRepository) *RawResponseIngest {
	return &RawResponseIngest{Responses: responses}
}

func (u *RawResponseIngest) Ingest(ctx context.Context, r domain.RawResponse) error {
	r.ReceivedAt = time.Now().UTC()
	return u.Responses.Save(ctx, r)
}

func (u *RawResponseIngest) List(ctx context.Context, limit int) ([]domain.RawResponse, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return u.Responses.List(ctx, limit)
}
