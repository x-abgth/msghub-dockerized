package repository

import (
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type AdminRepository interface {
	InsertAdminToDb(name, password string) error
	LoginAdmin(name, password string) (models.Admin, error)
	GetAdminsData(name string) ([]models.AdminModel, error)
	InsertAdminMessageDataRepository(message models.Message) error
	GetAllUsersDataForAdmin() ([]models.UserModel, error)
	GetGroupsData() ([]models.GroupModel, error)
	GetDeletedUserData() ([]models.UserModel, error)
	AdminBlockThisUserRepo(userID, duration string) error
	AdminBlockThisGroupRepo(groupID, duration string) error
}

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &repository{db}
}

func (r *repository) InsertAdminToDb(name, pass string) error {
	_, err2 := r.db.Exec(`INSERT INTO admins(admin_name, admin_pass) 
VALUES($1, $2);`, name, pass)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) LoginAdmin(uname, pass string) (models.Admin, error) {
	var name, password string
	rows, err := r.db.Query(
		`SELECT 
    	admin_name,
    	admin_pass
	FROM admins
	WHERE admin_name = $1;`, uname)

	if err != nil {
		return models.Admin{}, errors.New("an unknown error occurred, please try again")
	}

	defer rows.Close()
	for rows.Next() {
		if err1 := rows.Scan(
			&name,
			&password,
		); err1 != nil {
			return models.Admin{}, err1
		}
	}

	data := models.Admin{
		AdminName: name,
		AdminPass: password,
	}

	return data, nil
}

func (r *repository) GetAdminsData(uname string) ([]models.AdminModel, error) {
	var (
		adminID, adminName string
		res                []models.AdminModel
	)
	rows, err := r.db.Query(
		`SELECT 
		admin_id, 
    	admin_name
	FROM admins
	WHERE admin_name != $1;`, uname)

	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&adminID,
			&adminName,
		); err != nil {
			return res, err
		}

		data := models.AdminModel{
			AdminId:   adminID,
			AdminName: adminName,
		}

		res = append(res, data)
	}

	return res, nil
}

func (r *repository) InsertAdminMessageDataRepository(data models.Message) error {

	var (
		msgID int
		res   []int
	)

	rows, err := r.db.Query(
		`SELECT 
    	msg_id
	FROM admin_messages
	WHERE (is_recent = true) AND `, data.FromUserId, data.ToUserId, data.ToUserId, data.FromUserId)
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

	_, err2 := r.db.Exec(`INSERT INTO messages(from_user_id, to_user_id, content, sent_time, status, is_recent) 
VALUES($1, $2, $3, $4, $5, $6);`,
		data.FromUserId, data.ToUserId, data.Content, data.SentTime, data.Status, true)
	if err2 != nil {
		log.Println(err2)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) GetAllUsersDataForAdmin() ([]models.UserModel, error) {
	var (
		phone, name, about string
		avatar             *string
		isBlocked          bool
		res                []models.UserModel
	)
	rows, err := r.db.Query(
		`SELECT user_ph_no, user_name, user_avatar, user_about, is_blocked FROM users;`)
	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&phone,
			&name,
			&avatar,
			&about,
			&isBlocked,
		); err != nil {
			return res, err
		}

		null := ""
		if avatar == nil {
			avatar = &null
		}

		data := models.UserModel{
			UserPhone:     phone,
			UserAvatarUrl: *avatar,
			UserName:      name,
			UserAbout:     about,
			UserBlocked:   isBlocked,
		}

		res = append(res, data)
	}

	return res, nil
}

func (r *repository) GetGroupsData() ([]models.GroupModel, error) {
	var (
		id, name, about, date, members, creator string
		avatar                                  *string

		isBanned bool
		res      []models.GroupModel
	)
	rows, err := r.db.Query(
		`SELECT group_id, group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned FROM groups;`)
	if err != nil {
		return res, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&id,
			&name,
			&avatar,
			&about,
			&creator,
			&date,
			&members,
			&isBanned,
		); err != nil {
			return res, err
		}

		null := ""
		if avatar == nil {
			avatar = &null
		}

		m, err := strconv.Atoi(members)
		if err != nil {
			return res, err
		}

		data := models.GroupModel{
			Id:          id,
			Owner:       creator,
			Image:       *avatar,
			Name:        name,
			About:       about,
			CreatedDate: date,
			NoOfMembers: m,
			IsBanned:    isBanned,
		}

		res = append(res, data)
	}

	return res, nil
}

func (r *repository) GetDeletedUserData() ([]models.UserModel, error) {
	var (
		id, name, about, deleteTime string
		isBlocked                   bool
		avatar                      *string

		res []models.UserModel
	)
	rows, err := r.db.Query(
		`SELECT user_ph_no, user_avatar, user_about, is_blocked, delete_time FROM deleted_users;`)
	if err != nil {
		return nil, errors.New("an unknown error occurred, please try again")
	}

	for rows.Next() {
		if err := rows.Scan(
			&id,
			&name,
			&avatar,
			&about,
			&isBlocked,
			&deleteTime,
		); err != nil {
			return nil, err
		}

		null := ""
		if avatar == nil {
			avatar = &null
		}

		data := models.UserModel{
			UserPhone:     id,
			UserAvatarUrl: *avatar,
			UserName:      name,
			UserAbout:     about,
			IsBlocked:     isBlocked,
			DeletedTime:   deleteTime,
		}

		res = append(res, data)
	}

	return res, nil
}

func (r *repository) AdminBlockThisUserRepo(id, condition string) error {
	_, err1 := r.db.Exec(`UPDATE users SET is_blocked = true, blocked_duration = $1 WHERE user_ph_no = $2 AND is_blocked = false;`, condition, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) AdminBlockThisGroupRepo(id, condition string) error {
	_, err1 := r.db.Exec(`UPDATE groups SET is_banned = true, banned_time = $1 WHERE group_id = $2 AND is_banned = false;`, condition, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}
