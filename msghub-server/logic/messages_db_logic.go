package logic

import (
	"log"
	"os"
	"sort"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
)

type MessageDb struct {
	UserData repository.Message
}

// message status constants
const (
	IS_NOT_SENT  = "NOT_SENT"
	IS_SENT      = "SENT"
	IS_DELIVERED = "DELIVERED"
	IS_READ      = "READ"
)

const (
	TEXT  = "TEXT"
	IMAGE = "IMAGE"
)

// MigrateMessagesDb : Creates message table
func (m MessageDb) MigrateMessagesDb() error {
	err := m.UserData.CreateMessageTable()
	return err
}

func (m MessageDb) StorePersonalMessagesLogic(message models.MessageModel) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			os.Exit(1)
		}
	}()

	data := m.UserData
	data.Content = message.Content
	data.FromUserId = message.From
	data.ToUserId = message.To
	data.SentTime = message.Time
	data.ContentType = message.ContentType
	data.Status = message.Status

	err := m.UserData.InsertMessageDataRepository(data)
	if err != nil {
		panic(err.Error())
	}
}

func (m MessageDb) UpdatePmToDelivered(to string) error {
	return m.UserData.UpdateAllPersonalMessagesToDelivered(to)
}

func (m MessageDb) UpdatePmToRead(from, to string) error {
	return m.UserData.UpdateAllPersonalMessagesToRead(from, to)
}

func (m MessageDb) GetMessageDataLogic(target, from string) ([]models.MessageModel, error) {
	var this []models.MessageModel

	data1, err := m.UserData.GetAllPersonalMessages(from, target)
	if err != nil {
		return this, err
	}

	data2, err := m.UserData.GetAllPersonalMessages(target, from)
	if err != nil {
		return this, err
	}

	// Add admin messages also
	data1 = append(data1, data2...)

	for i := range data1 {
		myTime, err := time.Parse("2 Jan 2006 3:04:05 PM", data1[i].Time)
		if err != nil {
			return this, err
		}

		diff := time.Now().Sub(myTime)
		d := models.MessageModel{
			From:        data1[i].From,
			Content:     data1[i].Content,
			Time:        data1[i].Time,
			Status:      data1[i].Status,
			ContentType: data1[i].ContentType,
			Order:       float64(diff),
		}

		this = append(this, d)
	}

	sort.Slice(this, func(i, j int) bool {
		return this[i].Order > this[j].Order
	})

	return this, nil
}
