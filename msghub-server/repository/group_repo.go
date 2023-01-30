package repository

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type Group struct {
	GroupId           int    `json:"group_id"`
	GroupName         string `json:"group_name"`
	GroupAvatar       string `json:"group_avatar"`
	GroupAbout        string `json:"group_about"`
	GroupCreator      string `json:"group_creator"`
	GroupCreatedDate  string `json:"group_created_date"`
	GroupTotalMembers int    `json:"group_total_members"`
	IsBanned          bool   `json:"is_banned"`
	BannedTime        string `json:"banned_time"`
}

type UserGroupRelation struct {
	Id       int    `json:"id"`
	GroupId  int    `json:"group_id"`
	UserId   string `json:"user_id"`
	UserRole string `json:"user_role"`
}

type GroupMessage struct {
	MsgId          int    `json:"msg_id"`
	GroupId        int    `json:"group_id"`
	SenderId       string `json:"sender_id"`
	MessageContent string `json:"message_content"`
	ContentType    string `json:"content_type"`
	Status         string `json:"status"`
	SentTime       string `json:"sent_time"`
	IsRecent       bool   `json:"is_recent"`
}

func (group Group) CreateGroupTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS groups(group_id BIGSERIAL PRIMARY KEY NOT NULL, group_name TEXT NOT NULL, group_avatar TEXT NOT NULL, group_about TEXT NOT NULL, group_creator TEXT NOT NULL, group_created_date TEXT NOT NULL, group_total_members BIGINT NOT NULL, is_banned BOOLEAN NOT NULL, banned_time TEXT);`)
	return err
}

func (relation UserGroupRelation) CreateUserGroupRelationTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS user_group_relations(id BIGSERIAL PRIMARY KEY NOT NULL, group_id BIGINT NOT NULL, user_id TEXT NOT NULL, user_role TEXT NOT NULL);`)
	return err
}

func (gm GroupMessage) CreateGroupMessageTable() error {
	_, err := models.SqlDb.Exec(`CREATE TABLE IF NOT EXISTS group_messages(msg_id BIGSERIAL PRIMARY KEY NOT NULL, group_id BIGINT NOT NULL, sender_id TEXT NOT NULL, message_content TEXT NOT NULL, sent_time TEXT NOT NULL, is_recent BOOLEAN NOT NULL, status TEXT NOT NULL, content_type TEXT);`)
	return err
}

func CreateGroup(data Group) (int, error) {
	var id int
	if err := models.SqlDb.QueryRow(`INSERT INTO groups
		(group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned) 
VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING group_id`, data.GroupName, data.GroupAvatar, data.GroupAbout, data.GroupCreator, data.GroupCreatedDate, data.GroupTotalMembers, data.IsBanned).Scan(&id); err != nil {
		log.Println(err.Error())
		return 0, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return id, nil
}

func (group Group) GetGroupMemberCount(id string) string {
	var total string
	rows, err := models.SqlDb.Query(
		`SELECT 
    	group_total_members
	FROM groups
	WHERE group_id = $1`, id)
	if err != nil {
		return ""
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&total); err1 != nil {
			return ""
		}
	}

	return total
}

func (group Group) UpdateGroupMemberCount(num int, id string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE groups
		SET group_total_members = $1
		WHERE group_id = $2`, num, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (relation UserGroupRelation) CreateUserGroupRelation(groupId int, userId, role string) error {
	_, err1 := models.SqlDb.Exec(`INSERT INTO user_group_relations(
	                 group_id, user_id, user_role)
	VALUES($1, $2, $3);`,
		groupId, userId, role)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (gm GroupMessage) InsertGroupMessagesRepo(message GroupMessage) error {

	var (
		msgID int
		res   []int
	)

	rows, err := models.SqlDb.Query(
		`SELECT 
    	msg_id
	FROM group_messages
	WHERE (is_recent = true) AND group_id = $1`, message.GroupId)
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
		_, err1 := models.SqlDb.Exec(`UPDATE group_messages
		SET is_recent = false
		WHERE msg_id = $1`,
			res[i])
		if err1 != nil {
			log.Println(err1)
			return errors.New("sorry, An unknown error occurred. Please try again")
		}
	}

	_, err1 := models.SqlDb.Exec(`INSERT INTO group_messages(
	                 group_id, sender_id, message_content, content_type, status, sent_time, is_recent)
	VALUES($1, $2, $3, $4, $5, $6, $7);`,
		message.GroupId, message.SenderId, message.MessageContent, message.ContentType, message.Status, message.SentTime, true)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (relation UserGroupRelation) GetAllGroupsForAUser(ph string) ([]int, error) {
	var (
		num int
		res []int
	)
	rows, err := models.SqlDb.Query(
		`SELECT 
    	group_id
	FROM user_group_relations
	WHERE user_id = $1;`, ph)
	if err != nil {
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&num); err != nil {
			return res, err
		}
		res = append(res, num)
	}

	return res, nil
}

func (gm GroupMessage) GetRecentGroupMessages(id int) (models.GrpMsgModel, error) {
	var (
		groupID, name, avtr, sender, content, time string
		cType                                      *string
		res                                        models.GrpMsgModel
	)

	rows, err := models.SqlDb.Query(
		`SELECT groups.group_id, groups.group_name, groups.group_avatar, group_messages.sender_id, group_messages.message_content, group_messages.sent_time, group_messages.content_type 
FROM groups 
    INNER JOIN group_messages 
        ON groups.group_id = group_messages.group_id WHERE groups.group_id = $1 AND group_messages.is_recent = true ORDER BY sent_time;`, id)
	if err != nil {
		log.Println("From repo ===== ", err)
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&groupID,
			&name,
			&avtr,
			&sender,
			&content,
			&time,
			&cType,
		); err1 != nil {
			log.Println("The recent msg row from repo is ", err)
			return res, err1
		}

		null := ""
		if cType == nil {
			cType = &null
		}

		res = models.GrpMsgModel{
			Id:          groupID,
			Name:        name,
			Avatar:      avtr,
			Sender:      sender,
			Message:     content,
			ContentType: *cType,
			Time:        time,
		}
	}

	return res, nil
}

func (gm GroupMessage) GetAllMessagesFromGroup(id int) ([]models.MessageModel, error) {
	var (
		sender, message, time, cType string
		res                          []models.MessageModel
	)
	rows, err := models.SqlDb.Query(
		`SELECT sender_id, message_content, sent_time, content_type FROM group_messages WHERE group_id = $1;`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err1 := rows.Scan(
			&sender,
			&message,
			&time,
			&cType,
		); err1 != nil {
			return nil, err1
		}

		data := models.MessageModel{
			From:        sender,
			Content:     message,
			Time:        time,
			ContentType: strings.ToLower(cType),
		}

		res = append(res, data)
	}

	return res, nil
}

func (group Group) GetGroupDetailsRepo(id int) (models.GroupModel, error) {
	var (
		name, avatar, about, creator, date, totalMembers string
		banTime                                          *string
		isBan                                            bool
	)
	rows, err := models.SqlDb.Query(
		`SELECT group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned, banned_time FROM groups WHERE group_id = $1;`, id)
	if err != nil {
		return models.GroupModel{}, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&name,
			&avatar,
			&about,
			&creator,
			&date,
			&totalMembers,
			&isBan,
			&banTime,
		); err != nil {
			return models.GroupModel{}, err
		}
	}

	null := ""
	if banTime == nil {
		banTime = &null
	}

	n, nerr := strconv.Atoi(totalMembers)
	if nerr != nil {
		return models.GroupModel{}, nerr
	}
	return models.GroupModel{
		Name:        name,
		Image:       avatar,
		About:       about,
		Owner:       creator,
		CreatedDate: date,
		NoOfMembers: n,
		IsBanned:    isBan,
		BanTime:     *banTime,
	}, nil
}

func (group Group) CheckGroupBlockedRepo(id int) bool {
	var isBan bool
	rows, err := models.SqlDb.Query(
		`SELECT is_banned FROM groups WHERE group_id = $1;`, id)
	if err != nil {
		return false
	}

	defer rows.Close()

	for rows.Next() {
		if e := rows.Scan(&isBan); e != nil {
			return false
		}
	}

	return true
}

func (relation UserGroupRelation) GetAllTheGroupMembersRepo(id string) []string {
	var uid, role, admin string
	var res []string
	rows, err := models.SqlDb.Query(
		`SELECT user_id, user_role FROM user_group_relations WHERE group_id = $1 AND user_role != $2;`, id, "nil")
	if err != nil {
		return res
	}

	defer rows.Close()

	for rows.Next() {
		if e := rows.Scan(&uid, &role); e != nil {
			return res
		}

		if role == "admin" {
			admin = uid
			continue
		}
		res = append(res, uid)
	}

	if admin != "" {
		res = append([]string{admin}, res...)
	}

	return res
}

func (relation UserGroupRelation) IsUserGroupAdminRepo(gid, uid string) string {
	var role string
	rows, err := models.SqlDb.Query(
		`SELECT user_role FROM user_group_relations WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err != nil {
		return ""
	}

	defer rows.Close()

	for rows.Next() {
		if e := rows.Scan(&role); e != nil {
			return ""
		}
	}

	return role
}

func (relation UserGroupRelation) IsUserInGroupRepo(gid, uid string) string {
	var role string
	rows, err := models.SqlDb.Query(
		`SELECT user_role FROM user_group_relations WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err != nil {
		return ""
	}

	defer rows.Close()

	for rows.Next() {
		if e := rows.Scan(&role); e != nil {
			return ""
		}
	}

	return role
}

func (relation UserGroupRelation) UserGroupStatusUpdateRepo(gid, uid string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE user_group_relations
		SET user_role = 'member'
		WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (relation UserGroupRelation) UserLeftGroupRepo(gid, uid string) error {
	_, err1 := models.SqlDb.Exec(`UPDATE user_group_relations
		SET user_role = 'nil'
		WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}
