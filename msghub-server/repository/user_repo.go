package repository

import (
	"errors"
	"log"
	"strconv"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type User struct {
	UserPhNo        string  `json:"user_ph_no"`
	UserName        string  `json:"user_name"`
	UserAvatar      *string `json:"user_avatar"`
	UserAbout       string  `json:"user_about"`
	UserPassword    string  `json:"user_password"`
	IsBlocked       bool    `json:"is_blocked"`
	BlockedDuration *string `json:"blocked_duration"`
	BlockList       *string `json:"block_list"`
}

type DeletedUser struct {
	UserPhNo        string  `json:"user_ph_no"`
	UserAvatar      *string `json:"user_avatar"`
	UserAbout       string  `json:"user_about"`
	IsBlocked       bool    `json:"is_blocked"`
	BlockedDuration *string `json:"blocked_duration"`
	BlockList       *string `json:"block_list"`
	DeleteTime      string  `json:"delete_time"`
}

type Storie struct {
	UserId          string `json:"user_id"`
	StoryUrl        string `json:"story_url"`
	StoryUpdateTime string `json:"story_update_time"`
	Viewers         string `json:"viewers"`
	IsActive        bool   `json:"is_active"`
}

func (user User) CreateUserTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS users(user_ph_no TEXT PRIMARY KEY NOT NULL, user_name TEXT NOT NULL, user_avatar TEXT, user_about TEXT NOT NULL, user_password TEXT NOT NULL, is_blocked BOOLEAN NOT NULL, blocked_duration TEXT, block_list TEXT);`)
	return err
}

func (user User) CreateDeletedUserTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS deleted_users(user_ph_no TEXT PRIMARY KEY NOT NULL, user_avatar TEXT, user_about TEXT NOT NULL, is_blocked BOOLEAN NOT NULL, blocked_duration TEXT, block_list TEXT, delete_time TEXT);`)
	return err
}

func (user User) CreateStoiesTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS stories(user_id TEXT NOT NULL, story_url TEXT NOT NULL, story_update_time TEXT NOT NULL, viewers TEXT NOT NULL, is_active TEXT NOT NULL);`)
	return err
}

func (user User) GetUserDataUsingPhone(formPhone string) (int, models.UserModel, error) {
	var (
		model     models.UserModel
		userModel User
	)

	rows, err := models.SqlDb.Query(
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

func (user User) RegisterUser(formName, formPhone, formPass string) (bool, error) {
	defaultAbout := "Hey there! Send me a Hi."

	_, err1 := models.SqlDb.Exec(`INSERT INTO users(user_name, user_about, user_ph_no, user_password, is_blocked) 
VALUES($1, $2, $3, $4, $5);`,
		formName, defaultAbout, formPhone, formPass, false)
	if err1 != nil {
		log.Println(err1)
		return false, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return true, nil
}

func (user User) ReRegisterDeletedUser(phone, name, pass string) error {
	var (
		avatar, blockDur *string
		about, blockList string
		isBlocked        bool
	)

	_, err1 := models.SqlDb.Exec(`BEGIN TRANSACTION;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	rows, err := models.SqlDb.Query(
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

	_, err1 = models.SqlDb.Exec(`INSERT INTO users(user_ph_no, user_name, user_avatar, user_about, user_password, is_blocked, blocked_duration, block_list) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8);`, phone, name, *avatar, about, pass, isBlocked, *blockDur, blockList)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = models.SqlDb.Exec(`DELETE FROM deleted_users WHERE user_ph_no = $1`, phone)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = models.SqlDb.Exec(`COMMIT;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (user User) UserDuplicationStatus(phone string) int {
	var total = 0

	rows, err := models.SqlDb.Query(
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

func (user User) CheckDeletedUser(phone string) int {
	var total int
	rows, err := models.SqlDb.Query(
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

func (user User) GetUserData(ph string) (models.UserModel, error) {
	var name, phone, about string
	var isBlocked bool
	var avatar *string
	var data models.UserModel

	rows, err := models.SqlDb.Query(
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
func (user User) GetRecentChatList(ph string) ([]models.MessageModel, error) {
	var from, to, content, cType, sentTime, status string

	var res []models.MessageModel

	rows, err := models.SqlDb.Query(
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

func (user User) GetAllUsersData(ph string) ([]models.UserModel, error) {
	var name, phone, about string
	var avatar *string
	var res []models.UserModel

	rows, err := models.SqlDb.Query(
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

func (user User) GetGroupForUser(userId string) ([]int, error) {
	var group, userPh, role string
	rows, err := models.SqlDb.Query(
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

func (user User) AddStoryRepo(model Storie) error {
	_, err1 := models.SqlDb.Exec(`INSERT INTO stories(user_id, story_url, story_update_time, viewers, is_active) 
VALUES($1, $2, $3, $4, $5);`, model.UserId, model.StoryUrl, model.StoryUpdateTime, model.Viewers, model.IsActive)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) GetStoryViewersRepo(storyID string) string {
	var storyList string
	rows, err := models.SqlDb.Query(
		`SELECT 
    	viewers
	FROM stories
	WHERE user_id = $1;`, storyID)
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

func (user User) UpdateStoryViewersRepo(viewers, storyID string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE stories SET viewers = $1 WHERE user_id = $2`, viewers, storyID)
	if err1 != nil {
		log.Println(err1)

		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) DeleteStoryRepo(id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE stories SET is_active = false WHERE user_id = $1`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) CheckUserStory(userId string) (bool, int) {
	var (
		status bool
		count  int
	)
	rows, err := models.SqlDb.Query(
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

func (user User) UpdateStoryStatusRepo(url, time, uid string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE stories SET story_url = $1, story_update_time = $2, viewers = '', is_active = $3 WHERE user_id = $4;`, url, time, true, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (user User) GetAllUserStories() []Storie {
	var (
		res                    []Storie
		id, url, time, viewers string
	)

	rows, err := models.SqlDb.Query(
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

		data := Storie{
			UserId:          id,
			StoryUrl:        url,
			StoryUpdateTime: time,
			Viewers:         viewers,
		}

		res = append(res, data)
	}

	return res
}

func (user User) UpdateUserData(model models.UserModel) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET user_name = $1, user_about = $2, user_avatar = $3 WHERE user_ph_no = $4;`,
		model.UserName, model.UserAbout, model.UserAvatarUrl, model.UserPhone)
	if err1 != nil {
		return errors.New("couldn't execute the sql query")
	}

	return nil
}

func (user User) UndoAdminBlockRepo(id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET is_blocked = false, blocked_duration = '' WHERE user_ph_no = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) UnblockGroupRepo(id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE groups SET is_banned = false, banned_time = '' WHERE group_id = $1;`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (user User) GetUserBlockList(id string) (string, error) {
	var blockList *string
	rows, err := models.SqlDb.Query(
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

func (user User) DeleteUserAccountRepo(id, t string) error {
	var (
		name, about, blockList string
		avatar, blockDur       *string
		isBlocked              bool
	)
	_, err1 := models.SqlDb.Exec(`BEGIN TRANSACTION;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	rows, err := models.SqlDb.Query(
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

	_, err1 = models.SqlDb.Exec(`INSERT INTO deleted_users(user_ph_no, user_name, user_avatar, user_about, is_blocked, blocked_duration, block_list, delete_time) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8);`, id, name, *avatar, about, isBlocked, *blockDur, blockList, t)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = models.SqlDb.Exec(`DELETE FROM users WHERE user_ph_no = $1`, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	_, err1 = models.SqlDb.Exec(`COMMIT;`)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (user User) UpdateUserBlockList(id, val string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE users SET block_list = $1 WHERE user_ph_no = $2;`, val, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}
