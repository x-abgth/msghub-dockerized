package repository

import (
	"errors"
	"log"
	"strings"

	"github.com/x-abgth/msghub/msghub-server/models"
)

type Message struct {
	MsgId       int    `json:"msg_id"`
	FromUserId  string `json:"from_user_id"`
	ToUserId    string `json:"to_user_id"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	SentTime    string `json:"sent_time"`
	Status      string `json:"status"`
	IsRecent    bool   `json:"is_recent"`
}

func (m Message) CreateMessageTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS messages(msg_id BIGSERIAL PRIMARY KEY NOT NULL, from_user_id TEXT NOT NULL, to_user_id TEXT NOT NULL, content TEXT NOT NULL, content_type TEXT NOT NULL, sent_time TEXT NOT NULL, status TEXT NOT NULL, is_recent BOOLEAN NOT NULL);`)
	return err
}

func (m Message) InsertMessageDataRepository(data Message) error {

	var (
		msgID int
		res   []int
	)
	rows, err := models.SqlDb.Query(
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
		_, err1 := models.SqlDb.Exec(`UPDATE messages
		SET is_recent = false
		WHERE msg_id = $1`,
			res[i])
		if err1 != nil {
			log.Println(err1)
			return errors.New("sorry, An unknown error occurred. Please try again")
		}
	}

	_, err2 := models.SqlDb.Exec(`INSERT INTO messages(from_user_id, to_user_id, content, content_type, sent_time, status, is_recent) 
VALUES($1, $2, $3, $4, $5, $6, $7);`,
		data.FromUserId, data.ToUserId, data.Content, data.ContentType, data.SentTime, data.Status, true)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (m Message) UpdateAllPersonalMessagesToDelivered(to string) error {
	_, err2 := models.SqlDb.Exec(`UPDATE messages SET status = 'DELIVERED' WHERE to_user_id = $1 AND status = 'SENT';`, to)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (m Message) UpdateAllPersonalMessagesToRead(from, to string) error {
	_, err2 := models.SqlDb.Exec(`UPDATE messages SET status = 'READ' WHERE (from_user_id = $1 AND to_user_id = $2) AND status = 'DELIVERED';`, from, to)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (m Message) GetAllPersonalMessages(from, to string) ([]models.MessageModel, error) {

	var (
		fromID, msg, msgType, time, status string
		res                                []models.MessageModel
	)

	rows, err := models.SqlDb.Query(
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
