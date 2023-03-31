package logic

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
)

type GroupDataLogicModel struct {
	users             repository.User
	groupTb           repository.Group
	userGroupRelation repository.UserGroupRelation
	messageGroupTb    repository.GroupMessage
}

// MigrateUserDb :  Creates table for user according the struct User
func (group GroupDataLogicModel) MigrateGroupDb() error {
	err := group.groupTb.CreateGroupTable()
	return err
}

func (group GroupDataLogicModel) MigrateUserGroupDb() error {
	err := group.userGroupRelation.CreateUserGroupRelationTable()
	return err
}

func (group GroupDataLogicModel) MigrateGroupMessagesDb() error {
	err := group.messageGroupTb.CreateGroupMessageTable()
	return err
}

func (group GroupDataLogicModel) CreateGroupAndInsertDataLogic(groupData models.GroupModel) (bool, error) {
	// Get date of the group created
	t := time.Now()
	dateOfCreation := t.Format("02/01/2006")

	data := repository.Group{
		GroupName:         groupData.Name,
		GroupAvatar:       groupData.Image,
		GroupAbout:        groupData.About,
		GroupCreator:      groupData.Owner,
		GroupCreatedDate:  dateOfCreation,
		GroupTotalMembers: len(groupData.Members) + 1,
		IsBanned:          false,
	}

	id, err := repository.CreateGroup(data)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	err1 := group.userGroupRelation.CreateUserGroupRelation(id, groupData.Owner, "admin")
	if err != nil {
		log.Println(err1.Error())
		return false, err1
	}
	for i := range groupData.Members {
		err := group.userGroupRelation.CreateUserGroupRelation(id, groupData.Members[i], "member")
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
	}

	msg := models.GroupMessageModel{
		GroupId:  strconv.Itoa(id),
		SenderId: "admin",
		Content:  "+91 " + groupData.Owner + " created a group named " + groupData.Name + ".",
		Time:     time.Now().Format("2 Jan 2006 3:04:05 PM"),
	}
	err2 := group.InsertMessagesToGroup(msg)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}

func (group GroupDataLogicModel) AddGroupMembers(gid string, members []string) error {
	id, err := strconv.Atoi(gid)
	if err != nil {
		return err
	}
	for i := range members {
		msg := models.GroupMessageModel{
			GroupId:  gid,
			SenderId: "admin",
			Content:  "+91 " + members[i] + " has been added to the group.",
			Status:   "SENT",
			Time:     time.Now().Format("2 Jan 2006 3:04:05 PM"),
		}

		err = group.InsertMessagesToGroup(msg)
		if err != nil {
			return err
		}

		str := group.userGroupRelation.IsUserInGroupRepo(gid, members[i])
		if str == "nil" {
			err := group.userGroupRelation.UserGroupStatusUpdateRepo(gid, members[i])
			if err != nil {
				return err
			}
			continue
		}

		err := group.userGroupRelation.CreateUserGroupRelation(id, members[i], "member")
		if err != nil {
			return err
		}

	}

	val := group.groupTb.GetGroupMemberCount(gid)
	valI, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	valI += len(members)
	err = group.groupTb.UpdateGroupMemberCount(valI, gid)
	if err != nil {
		return err
	}

	return nil
}

func (group GroupDataLogicModel) InsertMessagesToGroup(message models.GroupMessageModel) error {
	var (
		err error
	)

	group.messageGroupTb.GroupId, err = strconv.Atoi(message.GroupId)
	if err != nil {
		return err
	}
	group.messageGroupTb.SenderId = message.SenderId
	group.messageGroupTb.MessageContent = message.Content
	group.messageGroupTb.ContentType = message.Type
	group.messageGroupTb.Status = message.Status
	group.messageGroupTb.SentTime = message.Time

	err1 := group.messageGroupTb.InsertGroupMessagesRepo(group.messageGroupTb)
	if err1 != nil {
		return err1
	}

	return nil
}

func (group GroupDataLogicModel) GetAllGroupMessagesLogic(groupID string) ([]models.MessageModel, error) {
	id, err := strconv.Atoi(groupID)
	if err != nil {
		return nil, err
	}

	data, err := group.messageGroupTb.GetAllMessagesFromGroup(id)
	if err != nil {
		return nil, err
	}

	// Need to sort the messages according to the time sent.
	for i := range data {
		messageSentTime, err := time.Parse("2 Jan 2006 3:04:05 PM", data[i].Time)
		if err != nil {
			return nil, err
		}
		diff := time.Now().Sub(messageSentTime)

		data[i].Order = float64(diff)
	}

	// Sorting the array of messages
	sort.Slice(data, func(i, j int) bool {
		return data[i].Order > data[j].Order
	})

	return data, nil
}

func (group GroupDataLogicModel) GetAllGroupMembersData(id string) []models.GroupMembersModel {
	// First get all the members id
	data := group.userGroupRelation.GetAllTheGroupMembersRepo(id)

	// Secondly get details of the members like - avatar, name, number
	var res []models.GroupMembersModel
	for i := range data {
		isAdmin := false
		if i == 0 {
			isAdmin = true
		} else {
			isAdmin = false
		}

		uData, err := group.users.GetUserData(data[i])
		if err != nil {
			return res
		}
		val := models.GroupMembersModel{
			MAvatar:  uData.UserAvatarUrl,
			MName:    uData.UserName,
			MPhone:   data[i],
			MIsAdmin: isAdmin,
		}

		res = append(res, val)
	}
	return res
}

func (group GroupDataLogicModel) GetGroupRecentChats(id int) (models.GrpMsgModel, error) {
	data, err := group.messageGroupTb.GetRecentGroupMessages(id)
	if err != nil {
		return models.GrpMsgModel{}, err
	}

	return data, nil
}

func (group GroupDataLogicModel) CheckUserLeftTheGroup(uid, gid string) bool {
	val := group.userGroupRelation.IsUserInGroupRepo(gid, uid)
	if val == "" || val == "nil" {
		return true
	}

	return false
}

func (group GroupDataLogicModel) GetGroupDetailsLogic(gId string) (models.GroupModel, error) {
	id, err := strconv.Atoi(gId)
	if err != nil {
		return models.GroupModel{}, err
	}

	data, err := group.groupTb.GetGroupDetailsRepo(id)
	return data, err
}

func (group GroupDataLogicModel) CheckUserIsInGroup(gId, uId string) bool {
	val := group.userGroupRelation.IsUserInGroupRepo(gId, uId)
	if val == "" || val == "nil" {
		return false
	}
	return true
}

func (group GroupDataLogicModel) CheckUserIsAdmin(gId, uId string) bool {
	role := group.userGroupRelation.IsUserGroupAdminRepo(gId, uId)
	if role == "admin" {
		return true
	}
	return false
}

func (group GroupDataLogicModel) UserLeftTheGroupLogic(groupId, userId, msg string) error {
	err := group.userGroupRelation.UserLeftGroupRepo(groupId, userId)
	if err != nil {
		return err
	}
	data := models.GroupMessageModel{
		GroupId:  groupId,
		SenderId: "admin",
		Content:  msg,
		Type:     "TEXT",
		Status:   "SENT",
		Time:     time.Now().Format("2 Jan 2006 3:04:05 PM"),
	}

	val := group.groupTb.GetGroupMemberCount(groupId)
	valI, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	valI--
	err = group.groupTb.UpdateGroupMemberCount(valI, groupId)
	if err != nil {
		return err
	}

	err = group.InsertMessagesToGroup(data)
	if err != nil {
		return err
	}

	return nil
}
