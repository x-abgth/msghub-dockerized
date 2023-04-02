package logic

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/repository"
	"github.com/x-abgth/msghub-dockerized/msghub-server/utils"
)

type UserLogic interface {
	UserLoginLogic(phone, password string) (models.UserModel, error)
	UserDuplicationStatsAndSendOtpLogic(phone string) error
	UserValidateOtpLogic(phone, otp string) error
	UserRegisterPhoneValidationLogic(phone string) error
	UserRegisterLogic(otp, name, phone, password string) bool
	GetDataForDashboardLogic(phone string) (models.UserDashboardModel, error)
	AddNewStoryLogic(userID, storyURL string) error
	StorySeenLogic(viewerID, storyID string) error
	DeleteUserStoryLogic(userID string) error
	GetAllUsersLogic(phone string) ([]models.UserModel, error)
	GetUserDataLogic(phone string) (models.UserModel, error)
	UpdateUserProfileDataLogic(name, about, image, phone string) error
	NonGroupMembersLogic(groupID, userID string) []models.UserModel
	GroupUnblockLogic(groupID string) error
	UserBlockUserLogic(userID, blockUserID string) error
	UserUnblockUserLogic(userID, blockedUserID string) error
	DeleteUserAccountLogic(userID string) error

	// For messages
	StorePersonalMessagesLogic(message models.MessageModel)
	UpdatePmToDelivered(toUserID string) error
	UpdatePmToRead(fromUserID, toUserID string) error
	GetMessageDataLogic(targetUserID, fromUserID string) ([]models.MessageModel, error)

	// For groups
	CreateGroupAndInsertDataLogic(groupModel models.GroupModel) (bool, error)
	AddGroupMembers(groupID string, groupMembers []string) error
	InsertMessagesToGroup(message models.GroupMessageModel) error
	GetAllGroupMessagesLogic(groupID string) ([]models.MessageModel, error)
	GetAllGroupMembersData(groupID string) []models.GroupMembersModel
	GetGroupRecentChats(groupID int) (models.GrpMsgModel, error)
	CheckUserLeftTheGroup(userID, groupID string) bool
	GetGroupDetailsLogic(groupID string) (models.GroupModel, error)
	CheckUserIsInGroup(groupID, userID string) bool
	CheckUserIsAdmin(groupID, userID string) bool
	UserLeftTheGroupLogic(groupID, userID, message string) error
}

type userDbLogic struct {
	userRepository    repository.UserRepository
	groupRepository   repository.GroupRepository
	messageRepository repository.MessageRepository
}

func NewUserLogic(userRepo repository.UserRepository, groupRepo repository.GroupRepository, messageRepo repository.MessageRepository) UserLogic {
	return &userDbLogic{userRepository: userRepo, groupRepository: groupRepo, messageRepository: messageRepo}
}

func (u *userDbLogic) UserLoginLogic(phone, password string) (models.UserModel, error) {
	var (
		count int
		user  models.UserModel
	)
	count, user, err := u.userRepository.GetUserDataUsingPhone(phone)
	if err != nil {
		return user, errors.New("you don't have an account, Please register")
	}

	// Check the value is isBlocked and if string convert to bool using if()
	if user.UserBlocked {
		if user.BlockDur == "permanent" {
			return user, errors.New("you have been permanently blocked from this website")
		} else {
			t, err := time.Parse("2-1-2006 3:04:05 PM", user.BlockDur)
			if err != nil {
				log.Println(err)
				return user, errors.New("an unknown error occurred, but you're blocked")
			}

			if float64(t.Sub(time.Now())) < 0.009 {
				err := u.userRepository.UndoAdminBlockRepo(phone)
				if err != nil {
					return user, errors.New("an unknown error occurred")
				}
				return user, nil
			}

			exp := strings.Split(user.BlockDur, " ")
			return user, errors.New("you have been blocked till " + exp[0] + " from this website")
		}
	} else if count < 1 {
		return user, errors.New("you don't have an account, Please register")
	} else if count > 1 {
		return user, errors.New("something went wrong. Try login again")
	} else {
		if utils.CheckPasswordMatch(password, user.UserPass) {
			return user, nil
		} else {
			return user, errors.New("invalid phone number or password")
		}
	}
}

func (u *userDbLogic) UserDuplicationStatsAndSendOtpLogic(phone string) error {

	var count int
	count, _, err := u.userRepository.GetUserDataUsingPhone(phone)
	if err != nil {
		return err
	}

	if count == 1 {
		status := utils.SendOtp(phone)
		if status {
			return nil
		} else {
			return errors.New("Couldn't send OTP to this number!")
		}
	} else {
		return errors.New("The phone number you entered is not registered!")
	}
}

func (u *userDbLogic) UserValidateOtpLogic(phone, otp string) error {
	status := utils.CheckOtp(phone, otp)
	if !status {
		return errors.New("The OTP is incorrect. Try again")
	}

	return nil
}

func (u *userDbLogic) UserRegisterPhoneValidationLogic(phone string) error {
	total := u.userRepository.UserDuplicationStatus(phone)
	if total > 0 {
		return errors.New("Account already exist with this number. Try Login method.")
	}
	fmt.Println("Sending otp to", phone)
	status := utils.SendOtp(phone)
	if status {
		return nil
	}
	return errors.New("Couldn't send OTP to this number. Please check the number.")
}

func (u *userDbLogic) UserRegisterLogic(otp, name, phone, pass string) bool {

	encPass, err := utils.HashEncrypt(pass)
	if err != nil {
		return false
	}

	status := utils.CheckOtp(phone, otp)
	if status {
		// Check user is in deleted table
		delU := u.userRepository.CheckDeletedUser(phone)
		if delU == 1 {
			// if yes, get user data
			err := u.userRepository.ReRegisterDeletedUser(name, phone, encPass)
			if err != nil {
				log.Println(err)
				return false
			}
			return true
		} else {
			done, _ := u.userRepository.RegisterUser(name, phone, encPass)
			if done {
				return true
			}
			return false
		}
	} else {
		return false
	}
}

func (u *userDbLogic) GetDataForDashboardLogic(phone string) (models.UserDashboardModel, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	blockList, err := u.userRepository.GetUserBlockList(phone)
	if err != nil {
		log.Println(err)
		return models.UserDashboardModel{}, err
	}

	blockArr := strings.Split(blockList, ",")

	// Got recent personal chats
	personalMessages, err := u.userRepository.GetRecentChatList(phone)
	if err != nil {
		log.Println(err)
		return models.UserDashboardModel{}, err
	}

	// Get recent group chats
	userGroups, err := u.userRepository.GetGroupForUser(phone)
	if err != nil {
		log.Println(err)
		return models.UserDashboardModel{}, err
	}

	// assign it to dashboard model
	var recents []models.RecentChatModel

	if len(personalMessages) > 0 {
		// Recent chats for personal messages
		for i := range personalMessages {
			var recentData models.RecentChatModel

			msgSentTime, err := time.Parse("2 Jan 2006 3:04:05 PM", personalMessages[i].Time)
			if err != nil {
				log.Println(err)
				break
			}
			diff := time.Now().Sub(msgSentTime)

			var pMsg string
			var pIsImage bool
			if personalMessages[i].ContentType == "IMAGE" {
				pMsg = "Image"
				pIsImage = true
			} else {
				pMsg = personalMessages[i].Content
				pIsImage = false
			}

			if personalMessages[i].From == phone {

				// Get user datas like dp, name
				_, usersData, err := u.userRepository.GetUserDataUsingPhone(personalMessages[i].To)
				if err != nil {
					log.Println(err)
					break
				}

				var isBlocked bool
				for j := range blockArr {
					if blockArr[j] == personalMessages[i].To {
						isBlocked = true
						break
					}
				}

				recentData = models.RecentChatModel{
					Content: models.RecentMessages{
						Id:          personalMessages[i].To,
						Name:        usersData.UserName,
						Avatar:      usersData.UserAvatarUrl,
						LastMsg:     pMsg,
						LastMsgTime: personalMessages[i].Time,
						IsImage:     pIsImage,
						IsRead:      false,
					},
					Sender:    personalMessages[i].From,
					IsGroup:   false,
					Order:     float64(diff),
					IsBlocked: isBlocked,
				}
			} else if personalMessages[i].From == "admin" {
				recentData = models.RecentChatModel{
					Content: models.RecentMessages{
						Id:          personalMessages[i].From,
						Name:        "🔴 ADMIN 🔴",
						Avatar:      "",
						LastMsg:     pMsg,
						LastMsgTime: personalMessages[i].Time,
						IsImage:     pIsImage,
					},
					Sender:    personalMessages[i].From,
					IsGroup:   false,
					Order:     float64(diff),
					IsBlocked: false,
					IsOnline:  false,
				}
			} else {
				// Get user datas like dp, name
				_, usersData, err := u.userRepository.GetUserDataUsingPhone(personalMessages[i].From)
				if err != nil {
					log.Println(err)
					break
				}

				var isBlocked bool
				for j := range blockArr {
					if blockArr[j] == personalMessages[i].From {
						isBlocked = true
						break
					}
				}

				var isRead bool
				if personalMessages[i].Status == IS_DELIVERED {
					isRead = true
				}

				recentData = models.RecentChatModel{
					Content: models.RecentMessages{
						Id:          personalMessages[i].From,
						Name:        usersData.UserName,
						Avatar:      usersData.UserAvatarUrl,
						LastMsg:     pMsg,
						LastMsgTime: personalMessages[i].Time,
						IsImage:     pIsImage,
						IsRead:      isRead,
					},
					Sender:    personalMessages[i].From,
					IsGroup:   false,
					Order:     float64(diff),
					IsBlocked: isBlocked,
				}
			}

			recents = append(recents, recentData)
		}
	}

	//  GETTING GROUP MESSAGES
	var groupMessages []models.GrpMsgModel

	for i := range userGroups {
		data, err := u.GetGroupRecentChats(userGroups[i])
		if err != nil {
			log.Println(err)
			break
		}
		if data.Id == "" {
			log.Println("The recent group message id is missing")
			continue
		}

		groupMessages = append(groupMessages, data)
	}

	if len(groupMessages) > 0 {
		for i := range groupMessages {
			groupSentTime, err := time.Parse("2 Jan 2006 3:04:05 PM", groupMessages[i].Time)
			if err != nil {
				log.Println(err)
				break
			}
			diff := time.Now().Sub(groupSentTime)

			var msg string
			var isImage bool
			if groupMessages[i].ContentType == "IMAGE" {
				msg = "Image"
				isImage = true
			} else {
				msg = groupMessages[i].Message
				isImage = false
			}

			recentData := models.RecentChatModel{
				Content: models.RecentMessages{
					Id:          groupMessages[i].Id,
					Name:        groupMessages[i].Name,
					Avatar:      groupMessages[i].Avatar,
					LastMsg:     msg,
					LastMsgTime: groupMessages[i].Time,
					IsImage:     isImage,
				},
				Sender:  groupMessages[i].Sender,
				IsGroup: true,
				Order:   float64(diff),
			}
			recents = append(recents, recentData)
		}
	}

	// sort the resultant array
	sort.Slice(recents, func(i, j int) bool {
		return recents[i].Order < recents[j].Order
	})

	// Get All stories
	data := u.userRepository.GetAllUserStories()

	var (
		storyModel []models.StoryModel
		userStory  models.StoryModel
	)

	// Get each user's name, avatar
	for i := range data {
		msgSentTime, err := time.Parse("2 Jan 2006 3:04:05 PM", data[i].StoryUpdateTime)
		if err != nil {
			log.Println(err)
			break
		}
		msgSentTime = msgSentTime.Add(time.Hour * 24)
		diff := float64(time.Now().Sub(msgSentTime))

		if diff <= 0 {
			dataX, err := u.GetUserDataLogic(data[i].UserId)
			if err != nil {
				log.Println(err)
				return models.UserDashboardModel{}, err
			}

			if dataX.UserPhone == phone {
				viwerStr := u.userRepository.GetStoryViewersRepo(phone)

				viewerArr := strings.Split(viwerStr, " ")

				var viewerModel []models.UserModel
				for i := range viewerArr {
					if len(viewerArr[i]) == 10 {
						z, err := u.GetUserDataLogic(viewerArr[i])
						if err != nil {
							log.Println(err)
							return models.UserDashboardModel{}, err
						}

						viewerModel = append(viewerModel, z)
					}
				}

				if len(viewerModel) < 1 {
					viewerModel = nil
				}

				y := models.StoryModel{
					UserName:    dataX.UserName,
					UserPhone:   dataX.UserPhone,
					UserAvatar:  dataX.UserAvatarUrl,
					StoryImg:    data[i].StoryUrl,
					Expiration:  data[i].StoryUpdateTime,
					ViewerCount: len(viewerModel),
					Viewers:     viewerModel,
				}
				userStory = y
			} else {
				viwerStr := u.userRepository.GetStoryViewersRepo(dataX.UserPhone)

				viewerArr := strings.Split(viwerStr, " ")

				var isViewed bool
				for i := range viewerArr {
					if viewerArr[i] == phone {
						isViewed = true
						break
					}
				}

				x := models.StoryModel{
					UserName:   dataX.UserName,
					UserPhone:  dataX.UserPhone,
					UserAvatar: dataX.UserAvatarUrl,
					StoryImg:   data[i].StoryUrl,
					Expiration: data[i].StoryUpdateTime,
					IsViewed:   isViewed,
				}
				storyModel = append(storyModel, x)
			}
		} else {
			// TODO: make story inactive
			err := u.userRepository.DeleteStoryRepo(data[i].UserId)
			if err != nil {
				log.Println(err)
				return models.UserDashboardModel{}, err
			}
			storyModel = nil
		}
	}

	// get user details
	userDetails, err1 := u.userRepository.GetUserData(phone)
	if err1 != nil {
		log.Println(err1)
		return models.UserDashboardModel{}, err1
	}

	if userStory.StoryImg == "" {
		return models.UserDashboardModel{
			UserPhone:      phone,
			UserDetails:    userDetails,
			RecentChatList: recents,
			StoryList:      storyModel,
		}, nil
	} else {
		return models.UserDashboardModel{
			UserPhone:      phone,
			UserDetails:    userDetails,
			UserStory:      userStory,
			RecentChatList: recents,
			StoryList:      storyModel,
		}, nil
	}
}

func (u *userDbLogic) AddNewStoryLogic(userId, story string) error {
	// Check if the user story exists in the database, if exists update the data
	status, count := u.userRepository.CheckUserStory(userId)
	if count == 1 && !status {
		// Update user story status
		err := u.userRepository.UpdateStoryStatusRepo(story, time.Now().Format("2 Jan 2006 3:04:05 PM"), userId)
		if err != nil {
			return err
		}

		return nil
	}

	// If the user story do not exist in the db
	data := models.Storie{
		UserId:          userId,
		StoryUrl:        story,
		StoryUpdateTime: time.Now().Format("2 Jan 2006 3:04:05 PM"),
		Viewers:         "",
		IsActive:        true,
	}
	err := u.userRepository.AddStoryRepo(data)
	if err != nil {
		return err
	}
	return nil
}

func (u *userDbLogic) StorySeenLogic(viewer, storyId string) error {
	// Get all viewers of the story
	sList := u.userRepository.GetStoryViewersRepo(storyId)

	var viewerList string

	if sList == "" {
		viewerList = viewer + " "
	} else {
		list := strings.Split(sList, " ")
		for i := range list {
			if len(list[i]) == 10 {
				if viewer != list[i] {
					viewerList = viewerList + list[i] + " "
				}
			}
		}
		viewerList = viewerList + viewer + " "
	}

	err := u.userRepository.UpdateStoryViewersRepo(viewerList, storyId)
	if err != nil {
		return err
	}

	return nil
}

func (u *userDbLogic) DeleteUserStoryLogic(userId string) error {
	err := u.userRepository.DeleteStoryRepo(userId)

	return err
}

func (u *userDbLogic) GetAllUsersLogic(ph string) ([]models.UserModel, error) {
	blockList, err := u.userRepository.GetUserBlockList(ph)
	if err != nil {
		log.Println(err)
		return []models.UserModel{}, err
	}

	blockArr := strings.Split(blockList, ",")

	data, err := u.userRepository.GetAllUsersDataForUser(ph)
	if err != nil {
		return []models.UserModel{}, err
	}

	var res []models.UserModel

	for i := range data {
		for j := range blockArr {
			if data[i].UserPhone == blockArr[j] {
				res = append(res, models.UserModel{
					UserName:      data[i].UserName,
					UserPhone:     data[i].UserPhone,
					UserAbout:     data[i].UserAbout,
					UserAvatarUrl: data[i].UserAvatarUrl,
					IsBlocked:     true,
				})
			} else {
				res = append(res, models.UserModel{
					UserName:      data[i].UserName,
					UserPhone:     data[i].UserPhone,
					UserAbout:     data[i].UserAbout,
					UserAvatarUrl: data[i].UserAvatarUrl,
					IsBlocked:     false,
				})
			}
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].UserName < res[j].UserName
	})

	return res, nil
}

func (u *userDbLogic) GetUserDataLogic(ph string) (models.UserModel, error) {
	data, err := u.userRepository.GetUserData(ph)
	if err != nil {
		return models.UserModel{}, err
	}
	return data, nil
}

func (u *userDbLogic) UpdateUserProfileDataLogic(name, about, image, phone string) error {
	data := models.UserModel{
		UserName:      name,
		UserAbout:     about,
		UserAvatarUrl: image,
		UserPhone:     phone,
	}

	err := u.userRepository.UpdateUserData(data)
	if err != nil {
		return err
	}
	return nil
}

func (u *userDbLogic) NonGroupMembersLogic(gid, uid string) []models.UserModel {
	// get all the users in the database
	uData, err := u.GetAllUsersLogic(uid)
	if err != nil {
		return nil
	}

	// get all the members in the group
	gData := u.GetAllGroupMembersData(gid)

	// filter them and get a slice of non-members
	var res []models.UserModel
	for i := range uData {
		flag := false
		for j := range gData {
			if uData[i].UserPhone == gData[j].MPhone {
				flag = !flag
			}
		}

		if flag == false {
			data := models.UserModel{
				UserAvatarUrl: uData[i].UserAvatarUrl,
				UserName:      uData[i].UserName,
				UserPhone:     uData[i].UserPhone,
				UserAbout:     uData[i].UserAbout,
			}

			res = append(res, data)
		}
	}

	return res
}

func (u *userDbLogic) GroupUnblockLogic(id string) error {
	err := u.userRepository.UnblockGroupRepo(id)

	return err
}

func (u *userDbLogic) UserBlockUserLogic(uid, buid string) error {
	list, err := u.userRepository.GetUserBlockList(uid)
	if err != nil {
		return err
	}

	if list == "" {
		list = buid
	} else {
		list = list + "," + buid
	}

	err = u.userRepository.UpdateUserBlockList(uid, list)
	if err != nil {
		return err
	}

	return nil
}

func (u *userDbLogic) UserUnblockUserLogic(uid, buid string) error {
	var val string

	list, err := u.userRepository.GetUserBlockList(uid)
	if err != nil {
		return err
	}

	if list == "" {
		return nil
	}
	data := strings.Split(list, ",")

	fmt.Println("Array of data = ", data)

	for i := range data {
		fmt.Println("The index of array = ", i)

		if data[i] == buid {
			continue
		} else {
			if val == "" {
				val = data[i]
			} else {
				val = val + "," + data[i]
			}
		}
		fmt.Println("Value = ", val)
	}

	err = u.userRepository.UpdateUserBlockList(uid, val)
	if err != nil {
		return err
	}

	return nil
}

func (u *userDbLogic) DeleteUserAccountLogic(id string) error {

	// Get the current time
	t := time.Now().Format("2 Jan 2006 3:04:05 PM")

	return u.userRepository.DeleteUserAccountRepo(id, t)
}
