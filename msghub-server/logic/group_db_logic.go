package logic

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

func (u *userDbLogic) CreateGroupAndInsertDataLogic(groupData models.GroupModel) (bool, error) {
	// Get date of the group created
	t := time.Now()
	dateOfCreation := t.Format("02/01/2006")

	data := models.Group{
		GroupName:         groupData.Name,
		GroupAvatar:       groupData.Image,
		GroupAbout:        groupData.About,
		GroupCreator:      groupData.Owner,
		GroupCreatedDate:  dateOfCreation,
		GroupTotalMembers: len(groupData.Members) + 1,
		IsBanned:          false,
	}

	id, err := u.groupRepository.CreateGroup(data)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	err1 := u.groupRepository.CreateUserGroupRelation(id, groupData.Owner, "admin")
	if err != nil {
		log.Println(err1.Error())
		return false, err1
	}
	for i := range groupData.Members {
		err := u.groupRepository.CreateUserGroupRelation(id, groupData.Members[i], "member")
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
	err2 := u.InsertMessagesToGroup(msg)
	if err2 != nil {
		return false, err2
	}

	return true, nil
}

func (u *userDbLogic) AddGroupMembers(gid string, members []string) error {
	id, err := strconv.Atoi(gid)
	if err != nil {
		return err
	}
	for i := range members {
		msg := models.GroupMessageModel{
			GroupId:  gid,
			SenderId: "admin",
			Content:  "+91 " + members[i] + " has been added to the u.groupRepository.",
			Status:   "SENT",
			Time:     time.Now().Format("2 Jan 2006 3:04:05 PM"),
		}

		err = u.InsertMessagesToGroup(msg)
		if err != nil {
			return err
		}

		str := u.groupRepository.IsUserInGroupRepo(gid, members[i])
		if str == "nil" {
			err := u.groupRepository.UserGroupStatusUpdateRepo(gid, members[i])
			if err != nil {
				return err
			}
			continue
		}

		err := u.groupRepository.CreateUserGroupRelation(id, members[i], "member")
		if err != nil {
			return err
		}

	}

	val := u.groupRepository.GetGroupMemberCount(gid)
	valI, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	valI += len(members)
	err = u.groupRepository.UpdateGroupMemberCount(valI, gid)
	if err != nil {
		return err
	}

	return nil
}

func (u *userDbLogic) InsertMessagesToGroup(message models.GroupMessageModel) error {
	var (
		err error
	)

	var grp models.GroupMessage

	grp.GroupId, err = strconv.Atoi(message.GroupId)
	if err != nil {
		return err
	}
	grp.SenderId = message.SenderId
	grp.MessageContent = message.Content
	grp.ContentType = message.Type
	grp.Status = message.Status
	grp.SentTime = message.Time

	err1 := u.groupRepository.InsertGroupMessagesRepo(grp)
	if err1 != nil {
		return err1
	}

	return nil
}

func (u *userDbLogic) GetAllGroupMessagesLogic(groupID string) ([]models.MessageModel, error) {
	id, err := strconv.Atoi(groupID)
	if err != nil {
		return nil, err
	}

	data, err := u.groupRepository.GetAllMessagesFromGroup(id)
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

func (u *userDbLogic) GetAllGroupMembersData(id string) []models.GroupMembersModel {
	// First get all the members id
	data := u.groupRepository.GetAllTheGroupMembersRepo(id)

	// Secondly get details of the members like - avatar, name, number
	var res []models.GroupMembersModel
	for i := range data {
		isAdmin := false
		if i == 0 {
			isAdmin = true
		} else {
			isAdmin = false
		}

		uData, err := u.userRepository.GetUserData(data[i])
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

func (u *userDbLogic) GetGroupRecentChats(id int) (models.GrpMsgModel, error) {
	data, err := u.groupRepository.GetRecentGroupMessages(id)
	if err != nil {
		return models.GrpMsgModel{}, err
	}

	return data, nil
}

func (u *userDbLogic) CheckUserLeftTheGroup(uid, gid string) bool {
	val := u.groupRepository.IsUserInGroupRepo(gid, uid)
	if val == "" || val == "nil" {
		return true
	}

	return false
}

func (u *userDbLogic) GetGroupDetailsLogic(gId string) (models.GroupModel, error) {
	id, err := strconv.Atoi(gId)
	if err != nil {
		return models.GroupModel{}, err
	}

	data, err := u.groupRepository.GetGroupDetailsRepo(id)
	return data, err
}

func (u *userDbLogic) CheckUserIsInGroup(gId, uId string) bool {
	val := u.groupRepository.IsUserInGroupRepo(gId, uId)
	if val == "" || val == "nil" {
		return false
	}
	return true
}

func (u *userDbLogic) CheckUserIsAdmin(gId, uId string) bool {
	role := u.groupRepository.IsUserGroupAdminRepo(gId, uId)
	if role == "admin" {
		return true
	}
	return false
}

func (u *userDbLogic) UserLeftTheGroupLogic(groupId, userId, msg string) error {
	err := u.groupRepository.UserLeftGroupRepo(groupId, userId)
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

	val := u.groupRepository.GetGroupMemberCount(groupId)
	valI, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	valI--
	err = u.groupRepository.UpdateGroupMemberCount(valI, groupId)
	if err != nil {
		return err
	}

	err = u.InsertMessagesToGroup(data)
	if err != nil {
		return err
	}

	return nil
}
