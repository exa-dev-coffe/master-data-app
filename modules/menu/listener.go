package menu

import (
	"encoding/json"
	"log/slog"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Listener interface {
}

type menuListener struct {
	service Service
	db      *sqlx.DB
	ch      *amqp.Channel
}

func NewListener(ch *amqp.Channel, db *sqlx.DB) Listener {
	repository := NewMenuRepository(db)
	uploadService := upload.NewUploadService()
	service := NewMenuService(repository, db, uploadService)

	l := &menuListener{service: service, db: db, ch: ch}

	if err := l.ListenSetRatingMenu(); err != nil {
		slog.Error("Failed to start listening to menu.set_rating queue", "error", err)
	}

	return l

}

func (l *menuListener) ListenSetRatingMenu() error {
	slog.Info("Starting to listen to menu.set_rating queue")
	return lib.ListenQueue(
		l.ch,
		"menu.set_rating",
		"",
		"menu.set_rating",
		lib.ExchangeDirect,
		func(delivery amqp.Delivery) error {
			slog.Info("Received message", "body", string(delivery.Body))
			var req UpdateRatingAndReviewCountRequest
			if err := json.Unmarshal(delivery.Body, &req); err != nil {
				slog.Error("Failed to parse message body", "error", err)
				return response.InternalServerError("Failed to parse message body", nil)
			}

			if err := lib.ValidateRequest(req); err != nil {
				return err
			}

			err := common.WithTransaction[UpdateRatingAndReviewCountRequest](l.db, l.service.UpdateRatingAndReviewCount, req)

			if err != nil {
				slog.Error("Failed to update menu rating and review count", "error", err)
				return response.InternalServerError("Failed to update menu rating and review count", nil)
			}

			return nil

		},
		true,
		false, false, false, false, "Master Data Update Rating Menu", true, amqp.Table{},
		amqp.Table{},
	)
}
