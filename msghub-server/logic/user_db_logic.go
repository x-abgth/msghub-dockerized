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

type UserDb struct {
	userData repository.User
	// groupMsg repository.GroupMessage
	// userData repository.UserRepository
	err error
}

// MigrateUserDb :  Creates table for user according the struct User
func (u *UserDb) MigrateUserDb() error {
	err := u.userData.CreateUserTable()
	return err
}

func (u *UserDb) MigrateDeletedUserDb() error {
	err := u.userData.CreateDeletedUserTable()
	return err
}

func (u *UserDb) MigrateStoriesDb() error {
	err := u.userData.CreateStoiesTable()
	return err
}

func (u *UserDb) UserLoginLogic(phone, password string) (bool, error) {

	var (
		count    int
		userData models.UserModel
	)
	count, userData, u.err = u.userData.GetUserDataUsingPhone(phone)
	if u.err != nil {
		return false, errors.New("you don't have an account, Please register")
	}

	// Check the value is isBlocked and if string convert to bool using if()
	if userData.UserBlocked {
		if userData.BlockDur == "permanent" {
			return false, errors.New("you have been permanently blocked from this website")
		} else {
			t, err := time.Parse("2-1-2006 3:04:05 PM", userData.BlockDur)
			if err != nil {
				log.Println(err)
				return false, errors.New("an unknown error occurred, but you're blocked")
			}

			if float64(t.Sub(time.Now())) < 0.009 {
				err := u.userData.UndoAdminBlockRepo(phone)
				if err != nil {
					return false, errors.New("an unknown error occurred")
				}
				return true, nil
			}

			exp := strings.Split(userData.BlockDur, " ")
			return false, errors.New("you have been blocked till " + exp[0] + " from this website")
		}
	} else if count < 1 {
		return false, errors.New("you don't have an account, Please register")
	} else if count > 1 {
		return false, errors.New("something went wrong. Try login again")
	} else {
		if utils.CheckPasswordMatch(password, userData.UserPass) {

			models.InitUserModel(userData)
			return true, nil
		} else {
			return false, errors.New("invalid phone number or password")
		}
	}
}

func (u *UserDb) UserDuplicationStatsAndSendOtpLogic(phone string) bool {

	var count int
	count, _, u.err = u.userData.GetUserDataUsingPhone(phone)
	if u.err != nil {
		return false
	}

	if count == 1 {
		status := utils.SendOtp(phone)
		if status {
			data := models.IncorrectOtpModel{
				PhoneNumber: phone,
				IsLogin:     true,
			}
			models.InitOtpErrorModel(data)
			return true
		} else {
			errorStr := models.IncorrectPhoneModel{
				ErrorStr: "Couldn't send OTP to this number!",
			}
			models.InitPhoneErrorModel(errorStr)
			return false
		}
	} else {
		errorStr := models.IncorrectPhoneModel{
			ErrorStr: "The phone number you entered is not registered!",
		}
		models.InitPhoneErrorModel(errorStr)
		return false
	}
}

func (u *UserDb) UserValidateOtpLogic(phone, otp string) bool {
	status := utils.CheckOtp(phone, otp)
	if status {
		return true
	} else {
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: phone,
			IsLogin:     true,
		}
		models.InitOtpErrorModel(data)
		return false
	}
}

func (u *UserDb) UserRegisterLogic(name, phone, pass string) bool {
	total := u.userData.UserDuplicationStatus(phone)
	encryptedFormPassword, err := utils.HashEncrypt(pass)

	if err != nil {
		alm := models.AuthErrorModel{
			ErrorStr: "The password is too weak. Please enter a strong password.",
		}
		models.InitAuthErrorModel(alm)
		return false
	} else {
		if total > 0 {
			alm := models.AuthErrorModel{
				ErrorStr: "Account already exist with this number. Try Login method.",
			}
			models.InitAuthErrorModel(alm)
			return false
		} else {
			fmt.Println("Sending otp to", phone)
			status := utils.SendOtp(phone)
			if status {
				data := models.IncorrectOtpModel{
					PhoneNumber: phone,
					IsLogin:     false,
				}
				user := models.UserModel{
					UserName:  name,
					UserPhone: phone,
					UserPass:  encryptedFormPassword,
				}
				models.InitUserModel(user)
				models.InitOtpErrorModel(data)
				return true
			} else {
				alm := models.AuthErrorModel{
					ErrorStr: "Couldn't send OTP to this number. Please check the number.",
				}
				models.InitAuthErrorModel(alm)
				return false
			}
		}
	}
}

func (u *UserDb) CheckUserRegisterOtpLogic(otp, name, phone, pass string) (bool, string) {
	status := utils.CheckOtp(phone, otp)
	if status {
		// Check user is in deleted table
		delU := u.userData.CheckDeletedUser(phone)
		if delU == 1 {
			// if yes, get user data
			err := u.userData.ReRegisterDeletedUser(phone, name, pass)
			if err != nil {
				log.Println(err)
				alm := models.AuthErrorModel{
					ErrorStr: err.Error(),
				}
				models.InitAuthErrorModel(alm)
				return false, "login"
			}
			return true, ""
		} else {
			done, alert := u.userData.RegisterUser(name, phone, pass)
			if done {
				return true, ""
			}
			alm := models.AuthErrorModel{
				ErrorStr: alert.Error(),
			}
			models.InitAuthErrorModel(alm)
			return false, "login"
		}
	} else {
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: phone,
			IsLogin:     false,
		}
		models.InitOtpErrorModel(data)
		return false, "otp"
	}
}

func (u *UserDb) GetDataForDashboardLogic(phone string) (models.UserDashboardModel, error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	blockList, err := u.userData.GetUserBlockList(phone)
	if err != nil {
		log.Println(err)
		return models.UserDashboardModel{}, err
	}

	blockArr := strings.Split(blockList, ",")

	// Got recent personal chats
	personalMessages, err := u.userData.GetRecentChatList(phone)
	if err != nil {
		log.Println(err)
		return models.UserDashboardModel{}, err
	}

	// Get recent group chats
	userGroups, err := u.userData.GetGroupForUser(phone)
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
				_, usersData, err := u.userData.GetUserDataUsingPhone(personalMessages[i].To)
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
				_, usersData, err := u.userData.GetUserDataUsingPhone(personalMessages[i].From)
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
	var grp GroupDataLogicModel

	for i := range userGroups {
		data, err := grp.GetGroupRecentChats(userGroups[i])
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
	data := u.userData.GetAllUserStories()

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
				viwerStr := u.userData.GetStoryViewersRepo(phone)

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
				viwerStr := u.userData.GetStoryViewersRepo(dataX.UserPhone)

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
			err := u.userData.DeleteStoryRepo(data[i].UserId)
			if err != nil {
				log.Println(err)
				return models.UserDashboardModel{}, err
			}
			storyModel = nil
		}
	}

	// get user details
	userDetails, err1 := u.userData.GetUserData(phone)
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

func (u *UserDb) AddNewStoryLogic(userId, story string) error {
	// Check if the user story exists in the database, if exists update the data
	status, count := u.userData.CheckUserStory(userId)
	if count == 1 && !status {
		// Update user story status
		err := u.userData.UpdateStoryStatusRepo(story, time.Now().Format("2 Jan 2006 3:04:05 PM"), userId)
		if err != nil {
			return err
		}

		return nil
	}

	// If the user story do not exist in the db
	data := repository.Storie{
		UserId:          userId,
		StoryUrl:        story,
		StoryUpdateTime: time.Now().Format("2 Jan 2006 3:04:05 PM"),
		Viewers:         "",
		IsActive:        true,
	}
	err := u.userData.AddStoryRepo(data)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserDb) StorySeenLogic(viewer, storyId string) error {
	// Get all viewers of the story
	sList := u.userData.GetStoryViewersRepo(storyId)

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

	err := u.userData.UpdateStoryViewersRepo(viewerList, storyId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserDb) DeleteUserStoryLogic(userId string) error {
	err := u.userData.DeleteStoryRepo(userId)

	return err
}

func (u *UserDb) GetAllUsersLogic(ph string) ([]models.UserModel, error) {
	blockList, err := u.userData.GetUserBlockList(ph)
	if err != nil {
		log.Println(err)
		return []models.UserModel{}, err
	}

	blockArr := strings.Split(blockList, ",")

	data, err := u.userData.GetAllUsersData(ph)
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

func (u *UserDb) GetUserDataLogic(ph string) (models.UserModel, error) {
	data, err := u.userData.GetUserData(ph)
	if err != nil {
		return models.UserModel{}, err
	}
	return data, nil
}

func (u *UserDb) UpdateUserProfileDataLogic(name, about, image, phone string) error {
	data := models.UserModel{
		UserName:      name,
		UserAbout:     about,
		UserAvatarUrl: image,
		UserPhone:     phone,
	}

	err := u.userData.UpdateUserData(data)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserDb) NonGroupMembersLogic(gid, uid string) []models.UserModel {
	// get all the users in the database
	uData, err := u.GetAllUsersLogic(uid)
	if err != nil {
		return nil
	}

	// get all the members in the group
	var gm GroupDataLogicModel
	gData := gm.GetAllGroupMembersData(gid)

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

func (u *UserDb) GroupUnblockLogic(id string) error {
	err := u.userData.UnblockGroupRepo(id)

	return err
}

func (u *UserDb) UserBlockUserLogic(uid, buid string) error {
	list, err := u.userData.GetUserBlockList(uid)
	if err != nil {
		return err
	}

	if list == "" {
		list = buid
	} else {
		list = list + "," + buid
	}

	err = u.userData.UpdateUserBlockList(uid, list)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserDb) UserUnblockUserLogic(uid, buid string) error {
	var val string

	list, err := u.userData.GetUserBlockList(uid)
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

	err = u.userData.UpdateUserBlockList(uid, val)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserDb) DeleteUserAccountLogic(id string) error {

	// Get the current time
	t := time.Now().Format("2 Jan 2006 3:04:05 PM")

	return u.userData.DeleteUserAccountRepo(id, t)
}
