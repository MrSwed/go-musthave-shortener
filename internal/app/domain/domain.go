package domain

type CreateURL struct {
	URL string `json:"url"`
}

type ResultURL struct {
	Result string `json:"result"`
}

type ShortBatchInputItem struct {
	CorrelationID string `json:"correlation_id" validate:"required"`
	OriginalURL   string `json:"original_url" validate:"required"`
}

type ShortBatchInput struct {
	List []ShortBatchInputItem `validate:"required,gt=0,dive"`
}
type ShortBatchResultItem struct {
	CorrelationTD string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
