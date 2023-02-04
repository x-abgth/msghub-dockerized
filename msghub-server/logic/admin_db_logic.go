package logic

import (
	"errors"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
)

type AdminDb struct {
	repo repository.Admin
	user repository.User
}

// MigrateAdminDb :  Creates table for admin according the struct Admin
func (admin AdminDb) MigrateAdminDb() error {
	err := admin.repo.CreateAdminTable()
	return err
}

func (admin AdminDb) InsertAdminLogic(username, password string) error {
	err := admin.repo.InsertAdminToDb(username, password)
	return err
}

func (admin AdminDb) AdminLoginLogic(username, password string) error {
	data, err := admin.repo.LoginAdmin(username, password)
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

func (admin AdminDb) GetAllAdminsData(name string) ([]models.AdminModel, error) {
	data, err := admin.repo.GetAdminsData(name)

	return data, err
}

func (admin AdminDb) GetUsersData() ([]models.UserModel, error) {
	data, err := admin.repo.GetAllUsersData()

	return data, err
}

func (admin AdminDb) GetDelUsersData() ([]models.UserModel, error) {
	data, err := admin.repo.GetDeletedUserData()

	return data, err
}

func (admin AdminDb) GetGroupsData() ([]models.GroupModel, error) {
	data, err := admin.repo.GetGroupsData()

	return data, err
}

func (admin AdminDb) BlockThisUserLogic(id, condition string) error {
	err := admin.repo.AdminBlockThisUserRepo(id, condition)

	return err
}

func (admin AdminDb) UnblockUserLogic(id string) error {
	err := admin.user.UndoAdminBlockRepo(id)

	return err
}

func (admin AdminDb) BlockThisGroupLogic(id, condition string) error {
	err := admin.repo.AdminBlockThisGroupRepo(id, condition)

	return err
}

func (admin AdminDb) AdminUnBlockGroupHandler(id string) error {
	err := admin.user.UnblockGroupRepo(id)

	return err
}
