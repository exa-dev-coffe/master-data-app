package menu

import (
	"encoding/json"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2/log"
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
		log.Fatalf("Failed to start listening to menu.set_rating queue: %v", err)
	}

	return l

}

func (l *menuListener) ListenSetRatingMenu() error {
	log.Info("Starting to listen to menu.set_rating queue")
	return lib.ListenQueue(
		l.ch,
		"menu.set_rating",
		"",
		"menu.set_rating",
		lib.ExchangeDirect,
		func(delivery amqp.Delivery) error {
			log.Info("Received message: ", string(delivery.Body))
			var req UpdateRatingAndReviewCountRequest
			if err := json.Unmarshal(delivery.Body, &req); err != nil {
				log.Errorf("Failed to parse message body: %v", err)
				return response.InternalServerError("Failed to parse message body", nil)
			}

			if err := lib.ValidateRequest(req); err != nil {
				return err
			}

			err := common.WithTransaction[UpdateRatingAndReviewCountRequest](l.db, l.service.UpdateRatingAndReviewCount, req)

			if err != nil {
				log.Errorf("Failed to update menu rating and review count: %v", err)
				return response.InternalServerError("Failed to update menu rating and review count", nil)
			}

			return nil

		},
		true,
		false, false, false, false, "Master Data Update Rating Menu", true, amqp.Table{},
		amqp.Table{},
	)
}
