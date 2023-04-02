package logic

import (
	"errors"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
)

type AdminLogic interface {
	InsertAdminLogic(username, password string) error
	AdminLoginLogic(username, password string) error
	GetAllAdminsData(username string) ([]models.AdminModel, error)
	GetUsersData() ([]models.UserModel, error)
	GetDelUsersData() ([]models.UserModel, error)
	GetGroupsData() ([]models.GroupModel, error)
	BlockThisUserLogic(userID, duration string) error
	UnblockUserLogic(userID string) error
	BlockThisGroupLogic(groupID, duration string) error
	AdminUnBlockGroupHandler(groupID string) error
	AdminStorePersonalMessages(message models.MessageModel) error
}

type adminDb struct {
	adminRepository   repository.AdminRepository
	userRepository    repository.UserRepository
	messageRepository repository.MessageRepository
}

func NewAdminLogic(adminRepo repository.AdminRepository, userRepo repository.UserRepository, messageRepo repository.MessageRepository) AdminLogic {
	return &adminDb{adminRepository: adminRepo, userRepository: userRepo, messageRepository: messageRepo}
}

func (a *adminDb) InsertAdminLogic(username, password string) error {
	err := a.adminRepository.InsertAdminToDb(username, password)
	return err
}

func (a *adminDb) AdminLoginLogic(username, password string) error {
	data, err := a.adminRepository.LoginAdmin(username, password)
	if err != nil {
		return err
	}

	if data.AdminName == username {
		if data.AdminPass == password {
			return nil
		}
		return errors.New("you have entered wrong password, please try again")
	}
	return errors.New("you have entered wrong password, please try again")
}

func (a *adminDb) GetAllAdminsData(name string) ([]models.AdminModel, error) {
	data, err := a.adminRepository.GetAdminsData(name)

	return data, err
}

func (a *adminDb) GetUsersData() ([]models.UserModel, error) {
	data, err := a.adminRepository.GetAllUsersDataForAdmin()

	return data, err
}

func (a *adminDb) GetDelUsersData() ([]models.UserModel, error) {
	data, err := a.adminRepository.GetDeletedUserData()

	return data, err
}

func (a *adminDb) GetGroupsData() ([]models.GroupModel, error) {
	data, err := a.adminRepository.GetGroupsData()

	return data, err
}

func (a *adminDb) BlockThisUserLogic(id, condition string) error {
	err := a.adminRepository.AdminBlockThisUserRepo(id, condition)

	return err
}

func (a *adminDb) UnblockUserLogic(id string) error {
	err := a.userRepository.UndoAdminBlockRepo(id)

	return err
}

func (a *adminDb) BlockThisGroupLogic(id, condition string) error {
	err := a.adminRepository.AdminBlockThisGroupRepo(id, condition)

	return err
}

func (a *adminDb) AdminUnBlockGroupHandler(id string) error {
	err := a.userRepository.UnblockGroupRepo(id)

	return err
}

func (a *adminDb) AdminStorePersonalMessages(message models.MessageModel) error {
	data := models.Message{
		Content:     message.Content,
		FromUserId:  message.From,
		ToUserId:    message.To,
		SentTime:    message.Time,
		ContentType: message.ContentType,
		Status:      message.Status,
	}

	err := a.messageRepository.InsertMessageDataRepository(data)
	if err != nil {
		return err
	}

	return nil
}
