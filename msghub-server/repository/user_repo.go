package repository

import (
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type UserRepository interface {
	// User table
	RegisterUser(name, phone, pass string) (bool, error)
	ReRegisterDeletedUser(name, phone, pass string) error
	GetUserDataUsingPhone(phone string) (int, models.UserModel, error)
	UserDuplicationStatus(userID string) int
	CheckDeletedUser(userID string) int
	GetUserData(userID string) (models.UserModel, error)
	GetRecentChatList(userID string) ([]models.MessageModel, error)
	GetAllUsersDataForUser(userID string) ([]models.UserModel, error)
	UpdateUserData(user models.UserModel) error
	UndoAdminBlockRepo(userID string) error
	GetUserBlockList(userID string) (string, error)
	DeleteUserAccountRepo(userID, deletedTime string) error
	UpdateUserBlockList(userID, blockList string) error
	// Storie table
	AddStoryRepo(story models.Storie) error
	GetStoryViewersRepo(userID string) string
	UpdateStoryViewersRepo(viewers, userID string) error
	DeleteStoryRepo(userID string) error
	CheckUserStory(userID string) (bool, int)
	UpdateStoryStatusRepo(storyURL, time, userID string) error
	GetAllUserStories() []models.Storie
	// Group table
	GetGroupForUser(userID string) ([]int, error)
	UnblockGroupRepo(groupID string) error
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &repository{db}
}

func (r *repository) GetUserDataUsingPhone(formPhone string) (int, models.UserModel, error) {
	var (
		model     models.UserModel
		userModel models.User
	)

	rows, err := r.db.Query(
		`SELECT 
    	user_avatar, 
    	user_about,
    	user_name, 
    	user_ph_no,
    	user_password, 
    	is_blocked, 
    	blocked_duration
	FROM users
	WHERE user_ph_no = $1;`, formPhone)
	if err != nil {
		return 0, models.UserModel{}, err
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&userModel.UserAvatar,
			&userModel.UserAbout,
			&userModel.UserName,
			&userModel.UserPhNo,
			&userModel.UserPassword,
			&userModel.IsBlocked,
			&userModel.BlockedDuration,
		); err1 != nil {
			return 0, models.UserModel{}, err1
		}
	}

	var blank = ""
	if userModel.UserAvatar == nil {
		userModel.UserAvatar = &blank
	}

	if userModel.BlockedDuration == nil {
		userModel.BlockedDuration = &blank
	}

	model = models.UserModel{
		UserAvatarUrl: *userModel.UserAvatar,
		UserAbout:     userModel.UserAbout,
		UserName:      userModel.UserName,
		UserPhone:     userModel.UserPhNo,
		UserPass:      userModel.UserPassword,
		UserBlocked:   userModel.IsBlocked,
		BlockDur:      *userModel.BlockedDuration,
	}

	return count, model, nil
}

func (r *repository) RegisterUser(formName, formPhone, formPass string) (bool, error) {
	defaultAbout := "Hey there! Send me a Hi."

	_, err1 := r.db.Exec(`INSERT INTO users(user_name, user_about, user_ph_no, user_password, is_blocked) 
VALUES($1, $2, $3, $4, $5);`,
		formName, defaultAbout, formPhone, formPass, false)
	if err1 != nil {
		log.Println(err1)
		return false, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return true, nil
}

func (r *repository) ReRegisterDeletedUser(name, phone, pass string) error {
	var (
		avatar, blockDur *string
		about, blockList string
		isBlocked        bool
	)

	_, err1 := r.db.Exec(`BEGIN TRANSACTION;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	rows, err := r.db.Query(
		`SELECT 
    	user_avatar, user_about, is_blocked, blocked_duration, block_list
	FROM deleted_users
	WHERE user_ph_no = $1;`, phone)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 = rows.Scan(
			&avatar,
			&about,
			&isBlocked,
			&blockDur,
			&blockList); err1 != nil {
			return err1
		}
	}

	null := ""
	if avatar == nil {
		avatar = &null
	}

	if blockDur == nil {
		blockDur = &null
	}

	_, err1 = r.db.Exec(`INSERT INTO users(user_ph_no, user_name, user_avatar, user_about, user_password, is_blocked, blocked_duration, block_list) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8);`, phone, name, *avatar, about, pass, isBlocked, *blockDur, blockList)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = r.db.Exec(`DELETE FROM deleted_users WHERE user_ph_no = $1`, phone)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = r.db.Exec(`COMMIT;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (r *repository) UserDuplicationStatus(phone string) int {
	var total = 0

	rows, err := r.db.Query(
		`SELECT *
	FROM users
	WHERE user_ph_no = $1;`, phone)
	if err != nil {
		log.Fatal("Error - ", err)
	}

	defer rows.Close()
	for rows.Next() {
		total++
	}

	return total
}

func (r *repository) CheckDeletedUser(phone string) int {
	var total int
	rows, err := r.db.Query(
		`SELECT *
	FROM deleted_users
	WHERE user_ph_no = $1;`, phone)
	if err != nil {
		log.Fatal("Error - ", err)
	}

	defer rows.Close()
	for rows.Next() {
		total++
	}
	return total
}

func (r *repository) GetUserData(ph string) (models.UserModel, error) {
	var name, phone, about string
	var isBlocked bool
	var avatar *string
	var data models.UserModel

	rows, err := r.db.Query(
		`SELECT 
    	user_avatar, 
    	user_name, 
    	user_ph_no,
    	user_about,
    	is_blocked
	FROM users
	WHERE user_ph_no = $1;`, ph)
	if err != nil {
		return data, err
	}

	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&avatar,
			&name,
			&phone,
			&about,
			&isBlocked); err1 != nil {
			return data, err1
		}
	}

	if avatar == nil {
		null := ""

		avatar = &null
	}
	data = models.UserModel{
		UserAvatarUrl: *avatar,
		UserName:      name,
		UserPhone:     phone,
		UserAbout:     about,
		UserBlocked:   isBlocked,
	}

	return data, nil
}

// This is not actual list, need to update
func (r *repository) GetRecentChatList(ph string) ([]models.MessageModel, error) {
	var from, to, content, cType, sentTime, status string

	var res []models.MessageModel

	rows, err := r.db.Query(
		`SELECT 
    from_user_id,
    	to_user_id, 
    	content, 
    	content_type,
    	sent_time,
    	status
	FROM messages
	WHERE is_recent = $1 AND ((from_user_id = $2 OR to_user_id = $3) OR (from_user_id = 'admin')) ORDER BY sent_time;`, true, ph, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&from,
			&to,
			&content,
			&cType,
			&sentTime,
			&status); err1 != nil {
			return res, err1
		}

		data := models.MessageModel{
			From:        from,
			To:          to,
			Content:     content,
			ContentType: cType,
			Time:        sentTime,
			Status:      status,
		}

		res = append(res, data)
	}
	return res, nil
}

func (r *repository) GetAllUsersDataForUser(ph string) ([]models.UserModel, error) {
	var name, phone, about string
	var avatar *string
	var res []models.UserModel

	rows, err := r.db.Query(
		`SELECT 
    	user_avatar, 
    	user_name, 
    	user_about,
    	user_ph_no 
	FROM users
	WHERE is_blocked = $1 AND user_ph_no != $2;`, false, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&avatar,
			&name,
			&about,
			&phone); err1 != nil {
			return res, err1
		}

		if avatar == nil {
			null := ""

			avatar = &null
		}
		data := models.UserModel{
			UserName:      name,
			UserPhone:     phone,
			UserAbout:     about,
			UserAvatarUrl: *avatar,
		}
		res = append(res, data)
	}

	return res, nil
}

func (r *repository) GetGroupForUser(userId string) ([]int, error) {
	var group, userPh, role string
	rows, err := r.db.Query(
		`SELECT 
    	group_id, 
    	user_id, 
    	user_role
	FROM user_group_relations
	WHERE user_id = $1;`, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []int
	for rows.Next() {
		if err1 := rows.Scan(
			&group,
			&userPh,
			&role); err1 != nil {
			return nil, err1
		}
		n, _ := strconv.Atoi(group)
		res = append(res, n)
	}
	return res, nil
}

func (r *repository) AddStoryRepo(model models.Storie) error {
	_, err1 := r.db.Exec(`INSERT INTO stories(user_id, story_url, story_update_time, viewers, is_active) 
VALUES($1, $2, $3, $4, $5);`, model.UserId, model.StoryUrl, model.StoryUpdateTime, model.Viewers, model.IsActive)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) GetStoryViewersRepo(userID string) string {
	var storyList string
	rows, err := r.db.Query(
		`SELECT 
    	viewers
	FROM stories
	WHERE user_id = $1;`, userID)
	if err != nil {
		return ""
	}

	defer rows.Close()
	for rows.Next() {
		if err1 := rows.Scan(&storyList); err1 != nil {
			return ""
		}
	}

	return storyList
}

func (r *repository) UpdateStoryViewersRepo(viewers, userID string) error {
	_, err1 := r.db.Exec(`UPDATE stories SET viewers = $1 WHERE user_id = $2`, viewers, userID)
	if err1 != nil {
		log.Println(err1)

		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) DeleteStoryRepo(id string) error {
	_, err1 := r.db.Exec(`UPDATE stories SET is_active = false WHERE user_id = $1`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) CheckUserStory(userId string) (bool, int) {
	var (
		status bool
		count  int
	)
	rows, err := r.db.Query(
		`SELECT 
    	is_active
	FROM stories
	WHERE user_id = $1;`, userId)
	if err != nil {
		return false, 0
	}

	defer rows.Close()

	for rows.Next() {
		count++
		if err1 := rows.Scan(
			&status); err1 != nil {
			return false, 0
		}
	}

	return status, count
}

func (r *repository) UpdateStoryStatusRepo(url, time, uid string) error {
	_, err1 := r.db.Exec(`UPDATE stories SET story_url = $1, story_update_time = $2, viewers = '', is_active = $3 WHERE user_id = $4;`, url, time, true, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (r *repository) GetAllUserStories() []models.Storie {
	var (
		res                    []models.Storie
		id, url, time, viewers string
	)

	rows, err := r.db.Query(
		`SELECT 
    	user_id, story_url, story_update_time, viewers 
	FROM stories
	WHERE is_active = $1;`, true)
	if err != nil {
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(&id, &url, &time, &viewers); err1 != nil {
			return nil
		}

		data := models.Storie{
			UserId:          id,
			StoryUrl:        url,
			StoryUpdateTime: time,
			Viewers:         viewers,
		}

		res = append(res, data)
	}

	return res
}

func (r *repository) UpdateUserData(model models.UserModel) error {
	_, err1 := r.db.Exec(`UPDATE users SET user_name = $1, user_about = $2, user_avatar = $3 WHERE user_ph_no = $4;`,
		model.UserName, model.UserAbout, model.UserAvatarUrl, model.UserPhone)
	if err1 != nil {
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (r *repository) UndoAdminBlockRepo(id string) error {
	_, err1 := r.db.Exec(`UPDATE users SET is_blocked = false, blocked_duration = '' WHERE user_ph_no = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) UnblockGroupRepo(id string) error {
	_, err1 := r.db.Exec(`UPDATE groups SET is_banned = false, banned_time = '' WHERE group_id = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) GetUserBlockList(id string) (string, error) {
	var blockList *string
	rows, err := r.db.Query(
		`SELECT 
    	block_list
	FROM users
	WHERE user_ph_no = $1;`, id)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&blockList); err1 != nil {
			return "", err1
		}
	}

	null := ""
	if blockList == nil {
		blockList = &null
	}

	return *blockList, nil
}

func (r *repository) DeleteUserAccountRepo(id, t string) error {
	var (
		name, about, blockList string
		avatar, blockDur       *string
		isBlocked              bool
	)
	_, err1 := r.db.Exec(`BEGIN TRANSACTION;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	rows, err := r.db.Query(
		`SELECT 
    	user_name, user_avatar, user_about, is_blocked, blocked_duration, block_list
	FROM users
	WHERE user_ph_no = $1;`, id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 = rows.Scan(
			&name,
			&avatar,
			&about,
			&isBlocked,
			&blockDur,
			&blockList); err1 != nil {
			return err1
		}
	}

	null := ""
	if avatar == nil {
		avatar = &null
	}
	if blockDur == nil {
		blockDur = &null
	}

	_, err1 = r.db.Exec(`INSERT INTO deleted_users(user_ph_no, user_name, user_avatar, user_about, is_blocked, blocked_duration, block_list, delete_time) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8);`, id, name, *avatar, about, isBlocked, *blockDur, blockList, t)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = r.db.Exec(`DELETE FROM users WHERE user_ph_no = $1`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = r.db.Exec(`COMMIT;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (r *repository) UpdateUserBlockList(id, val string) error {
	_, err1 := r.db.Exec(`UPDATE users SET block_list = $1 WHERE user_ph_no = $2;`, val, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}
