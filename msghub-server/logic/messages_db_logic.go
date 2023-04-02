package logic

import (
	"log"
	"os"
	"sort"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

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

func (u *userDbLogic) StorePersonalMessagesLogic(message models.MessageModel) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			os.Exit(1)
		}
	}()

	data := models.Message{
		Content:     message.Content,
		FromUserId:  message.From,
		ToUserId:    message.To,
		SentTime:    message.Time,
		ContentType: message.ContentType,
		Status:      message.Status,
	}

	err := u.messageRepository.InsertMessageDataRepository(data)
	if err != nil {
		panic(err.Error())
	}
}

func (u *userDbLogic) UpdatePmToDelivered(to string) error {
	return u.messageRepository.UpdateAllPersonalMessagesToDelivered(to)
}

func (u *userDbLogic) UpdatePmToRead(from, to string) error {
	return u.messageRepository.UpdateAllPersonalMessagesToRead(from, to)
}

func (u *userDbLogic) GetMessageDataLogic(target, from string) ([]models.MessageModel, error) {
	var this []models.MessageModel

	data1, err := u.messageRepository.GetAllPersonalMessages(from, target)
	if err != nil {
		return this, err
	}

	data2, err := u.messageRepository.GetAllPersonalMessages(target, from)
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
