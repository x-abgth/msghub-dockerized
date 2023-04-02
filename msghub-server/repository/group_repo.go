package repository

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

type GroupRepository interface {
	CreateGroup(group models.Group) (int, error)
	GetGroupMemberCount(groupID string) string
	UpdateGroupMemberCount(count int, groupID string) error
	CreateUserGroupRelation(groupID int, userID, userRole string) error
	InsertGroupMessagesRepo(message models.GroupMessage) error
	GetAllGroupsForAUser(userID string) ([]int, error)
	GetRecentGroupMessages(groupID int) (models.GrpMsgModel, error)
	GetAllMessagesFromGroup(groupID int) ([]models.MessageModel, error)
	GetGroupDetailsRepo(groupID int) (models.GroupModel, error)
	CheckGroupBlockedRepo(groupID int) bool
	GetAllTheGroupMembersRepo(groupID string) []string
	IsUserGroupAdminRepo(groupID, userID string) string
	IsUserInGroupRepo(groupID, userID string) string
	UserGroupStatusUpdateRepo(groupID, userID string) error
	UserLeftGroupRepo(groupID, userID string) error
}

func NewGroupRepository(db *sql.DB) GroupRepository {
	return &repository{db}
}

func (r *repository) CreateGroup(data models.Group) (int, error) {
	var id int
	if err := r.db.QueryRow(`INSERT INTO groups
		(group_name, group_avatar, group_about, group_creator, group_created_date, group_total_members, is_banned) 
VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING group_id`, data.GroupName, data.GroupAvatar, data.GroupAbout, data.GroupCreator, data.GroupCreatedDate, data.GroupTotalMembers, data.IsBanned).Scan(&id); err != nil {
		log.Println(err.Error())
		return 0, errors.New("sorry, An unknown error occurred. Please try again")
	}

	return id, nil
}

func (r *repository) GetGroupMemberCount(id string) string {
	var total string
	rows, err := r.db.Query(
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

func (r *repository) UpdateGroupMemberCount(num int, id string) error {
	_, err1 := r.db.Exec(`UPDATE groups
		SET group_total_members = $1
		WHERE group_id = $2`, num, id)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (r *repository) CreateUserGroupRelation(groupId int, userId, role string) error {
	_, err1 := r.db.Exec(`INSERT INTO user_group_relations(
	                 group_id, user_id, user_role)
	VALUES($1, $2, $3);`,
		groupId, userId, role)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) InsertGroupMessagesRepo(message models.GroupMessage) error {

	var (
		msgID int
		res   []int
	)

	rows, err := r.db.Query(
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
		_, err1 := r.db.Exec(`UPDATE group_messages
		SET is_recent = false
		WHERE msg_id = $1`,
			res[i])
		if err1 != nil {
			log.Println(err1)
			return errors.New("sorry, An unknown error occurred. Please try again")
		}
	}

	_, err1 := r.db.Exec(`INSERT INTO group_messages(
	                 group_id, sender_id, message_content, content_type, status, sent_time, is_recent)
	VALUES($1, $2, $3, $4, $5, $6, $7);`,
		message.GroupId, message.SenderId, message.MessageContent, message.ContentType, message.Status, message.SentTime, true)
	if err1 != nil {
		log.Println(err1.Error())
		return errors.New("sorry, An unknown error occurred. Please try again")
	}

	return nil
}

func (r *repository) GetAllGroupsForAUser(ph string) ([]int, error) {
	var (
		num int
		res []int
	)
	rows, err := r.db.Query(
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

func (r *repository) GetRecentGroupMessages(id int) (models.GrpMsgModel, error) {
	var (
		groupID, name, avtr, sender, content, time string
		cType                                      *string
		res                                        models.GrpMsgModel
	)

	rows, err := r.db.Query(
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

func (r *repository) GetAllMessagesFromGroup(id int) ([]models.MessageModel, error) {
	var (
		sender, message, time, cType string
		res                          []models.MessageModel
	)
	rows, err := r.db.Query(
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

func (r *repository) GetGroupDetailsRepo(id int) (models.GroupModel, error) {
	var (
		name, avatar, about, creator, date, totalMembers string
		banTime                                          *string
		isBan                                            bool
	)
	rows, err := r.db.Query(
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

func (r *repository) CheckGroupBlockedRepo(id int) bool {
	var isBan bool
	rows, err := r.db.Query(
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

func (r *repository) GetAllTheGroupMembersRepo(id string) []string {
	var uid, role, admin string
	var res []string
	rows, err := r.db.Query(
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

func (r *repository) IsUserGroupAdminRepo(gid, uid string) string {
	var role string
	rows, err := r.db.Query(
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

func (r *repository) IsUserInGroupRepo(gid, uid string) string {
	var role string
	rows, err := r.db.Query(
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

func (r *repository) UserGroupStatusUpdateRepo(gid, uid string) error {
	_, err1 := r.db.Exec(`UPDATE user_group_relations
		SET user_role = 'member'
		WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}

func (r *repository) UserLeftGroupRepo(gid, uid string) error {
	_, err1 := r.db.Exec(`UPDATE user_group_relations
		SET user_role = 'nil'
		WHERE group_id = $1 AND user_id = $2;`, gid, uid)
	if err1 != nil {
		log.Println(err1)
		return errors.New("sorry, An unknown error occurred. Please try again")
	}
	return nil
}
