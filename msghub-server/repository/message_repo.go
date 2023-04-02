package repository

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type MessageRepository interface {
	InsertMessageDataRepository(message models.Message) error
	UpdateAllPersonalMessagesToDelivered(toUserID string) error
	UpdateAllPersonalMessagesToRead(fromUserID, toUserID string) error
	GetAllPersonalMessages(fromUserID, toUserID string) ([]models.MessageModel, error)
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &repository{db}
}

func (r *repository) InsertMessageDataRepository(data models.Message) error {

	var (
		msgID int
		res   []int
	)
	rows, err := r.db.Query(
		`SELECT 
    	msg_id
	FROM messages
	WHERE (is_recent = true) AND (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $3 AND to_user_id = $4);`, data.FromUserId, data.ToUserId, data.ToUserId, data.FromUserId)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&msgID); err1 != nil {
			return err1
		}

		res = append(res, msgID)
	}

	for i := range res {
		_, err1 := r.db.Exec(`UPDATE messages
		SET is_recent = false
		WHERE msg_id = $1`,
			res[i])
		if err1 != nil {
			log.Println(err1)
			return errors.New("sorry, An unknown error occurred. Please try again")
		}
	}

	_, err2 := r.db.Exec(`INSERT INTO messages(from_user_id, to_user_id, content, content_type, sent_time, status, is_recent) 
VALUES($1, $2, $3, $4, $5, $6, $7);`,
		data.FromUserId, data.ToUserId, data.Content, data.ContentType, data.SentTime, data.Status, true)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) UpdateAllPersonalMessagesToDelivered(to string) error {
	_, err2 := r.db.Exec(`UPDATE messages SET status = 'DELIVERED' WHERE to_user_id = $1 AND status = 'SENT';`, to)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) UpdateAllPersonalMessagesToRead(from, to string) error {
	_, err2 := r.db.Exec(`UPDATE messages SET status = 'READ' WHERE (from_user_id = $1 AND to_user_id = $2) AND status = 'DELIVERED';`, from, to)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) GetAllPersonalMessages(from, to string) ([]models.MessageModel, error) {

	var (
		fromID, msg, msgType, time, status string
		res                                []models.MessageModel
	)

	rows, err := r.db.Query(
		`SELECT 
    	from_user_id, 
    	content,
    	content_type,
    	sent_time,
    	status
	FROM messages
	WHERE from_user_id = $1 AND to_user_id = $2;`, from, to)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&fromID,
			&msg,
			&msgType,
			&time,
			&status); err1 != nil {
			return res, err1
		}

		data := models.MessageModel{
			From:        fromID,
			Content:     msg,
			ContentType: strings.ToLower(msgType),
			Time:        time,
			Status:      status,
		}
		res = append(res, data)
	}

	return res, nil
}
