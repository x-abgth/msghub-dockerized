package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/x-abgth/msghub/msghub-server/models"
	"github.com/x-abgth/msghub/msghub-server/template"
	"github.com/x-abgth/msghub/msghub-server/utils"
	jwtPkg "github.com/x-abgth/msghub/msghub-server/utils/jwt"

	"github.com/x-abgth/msghub/msghub-server/logic"

	"github.com/gorilla/mux"

	"net/http"
)

// This might helps to pass error strings from one route to other
type InformationHelper struct {
	userRepo     logic.UserDb
	messagesRepo logic.MessageDb
	groupRepo    logic.GroupDataLogicModel
	errorStr     string
}

func (info *InformationHelper) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := info.userRepo.MigrateUserDb()
	if err != nil {
		log.Fatal("Error creating user table : ", err.Error())
	}

	err = info.userRepo.MigrateDeletedUserDb()
	if err != nil {
		log.Fatal("Error creating deleted user table : ", err.Error())
	}

	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)

	http.SetCookie(w, &http.Cookie{
		Name:     "userToken",
		Value:    token,
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	alm := struct {
		ErrorStr string
	}{
		ErrorStr: info.errorStr,
	}
	err1 := template.Tpl.ExecuteTemplate(w, "index.html", alm)
	if err1 != nil {
		fmt.Println("Error : ", err1.Error())
	}
}

func (info *InformationHelper) UserLoginCredentialsHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	r.ParseForm()

	ph := r.PostFormValue("signinPh")
	pass := r.PostFormValue("signinPass")

	handleExceptions(w, r, "/")

	isValid, alert := info.userRepo.UserLoginLogic(ph, pass)

	if isValid {
		data := models.ReturnUserModel()
		// assigning JWT tokens
		claims := &jwtPkg.UserJwtClaim{
			User:            *data,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		info.errorStr = alert.Error()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// This handler displays the page to enter the phone number
func (info *InformationHelper) UserLoginWithOtpPhonePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	data := models.ReturnOtpErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "login_with_otp.html", data)
	if err != nil {
		panic(err)
	}
}

// This handler process the phone number given and check weather is valid or not
func (info *InformationHelper) UserLoginOtpPhoneValidateHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()
	r.ParseForm()

	ph := r.PostFormValue("phone")

	if user := info.userRepo.UserDuplicationStatsAndSendOtpLogic(ph); user {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login/otp/getphone", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserOtpPageHandler(w http.ResponseWriter, r *http.Request) {
	data := models.ReturnOtpErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "user_otp_validation.html", data)
	utils.PrintError(err, "")
}

func (info *InformationHelper) UserVerifyLoginOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	otp := r.PostFormValue("loginOtp")
	ph := r.PostFormValue("loginPhone")

	status := info.userRepo.UserValidateOtpLogic(ph, otp)

	if status {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
	}
}

func (info *InformationHelper) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.PostFormValue("signupName")
	ph := r.PostFormValue("signupPh")
	pass := r.PostFormValue("signupPass")

	// This only sends the otp for registration, won't update anything on db
	status := info.userRepo.UserRegisterLogic(name, ph, pass)

	if status {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/register/otp/getotp", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserVerifyRegisterOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	otp := r.PostFormValue("loginOtp")

	user := models.ReturnUserModel()

	if ok, flag := info.userRepo.CheckUserRegisterOtpLogic(otp, user.UserName, user.UserPhone, user.UserPass); ok {
		userData := models.UserModel{
			UserName:  user.UserName,
			UserPhone: user.UserPhone,
		}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		if flag == "login" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else if flag == "otp" {
			http.Redirect(w, r, "/register/otp/getotp", http.StatusSeeOther)
		}
	}
}

func (info *InformationHelper) UserDashboardHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			log.Println(e)
			cookie := &http.Cookie{Name: "userToken", MaxAge: -1, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	// Creates table for user stories
	storyErr := info.userRepo.MigrateStoriesDb()
	if storyErr != nil {
		panic(storyErr.Error())
	}

	// Creates table for user's personal messages
	migrateErr := info.messagesRepo.MigrateMessagesDb()
	if migrateErr != nil {
		panic(migrateErr.Error())
	}

	// Group migration statements
	groupMigrationError := info.groupRepo.MigrateGroupDb()
	if groupMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupMigrationError.Error())
	}

	groupUserMigrationError := info.groupRepo.MigrateUserGroupDb()
	if groupUserMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupUserMigrationError.Error())
	}

	groupMessageMigrationError := info.groupRepo.MigrateGroupMessagesDb()
	if groupMessageMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupMessageMigrationError.Error())
	}

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		log.Println("NO cookie")
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	// Update every sent status of the user to delivered, when the user gets online
	err := info.messagesRepo.UpdatePmToDelivered(claim.User.UserPhone)
	if err != nil {
		panic(err)
	}

	data, err := info.userRepo.GetDataForDashboardLogic(claim.User.UserPhone)
	if err != nil {
		log.Println("Error getting dashboard logic")
		info.errorStr = err.Error()
		panic(err.Error())
	}

	info.errorStr = ""
	err2 := template.Tpl.ExecuteTemplate(w, "user_dashboard.html", data)
	if err2 != nil {
		log.Println(err2)
		panic("Not yet in dashboard page")
	}
}

func (info *InformationHelper) UserAddStoryHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	file, _, _ := r.FormFile("story_photo")

	var imageNameA string
	if file != nil {
		// Check weather the string is same as in db
		imageNameA = utils.StoreThisFileInBucket("stories/", target+"-story", file)
		defer file.Close()
	}

	// Update data to the database (first create the table for the story)
	err := info.userRepo.AddNewStoryLogic(target, imageNameA)
	if err != nil {
		panic(err)
	}

	// redirect to the dashboard
	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserStorySeenHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(e)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		log.Println("NO cookie")
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	// Add story viewers
	err := info.userRepo.StorySeenLogic(claim.User.UserPhone, target)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(nil)
}

func (info *InformationHelper) UserDeleteStoryHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		log.Println("No cookie")
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	if target != claim.User.UserPhone {
		panic("Invalid access to delete story")
	}

	// Delete story
	err := info.userRepo.DeleteUserStoryLogic(target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			log.Println(e)
			cookie := &http.Cookie{Name: "userToken", MaxAge: -1, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	// Take values from the database
	userInfo, err2 := info.userRepo.GetUserDataLogic(claim.User.UserPhone)
	if err2 != nil {
		panic(err2.Error())
	}

	data := struct {
		Name  string
		About string
		Phone string
		Image string
	}{
		Name:  userInfo.UserName,
		About: userInfo.UserAbout,
		Phone: userInfo.UserPhone,
		Image: userInfo.UserAvatarUrl,
	}

	err := template.Tpl.ExecuteTemplate(w, "user_profile_update.html", data)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println("ERROR HAPPENED -- ", e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	err := r.ParseMultipartForm(10 << 24)
	if err != nil {
		panic(err.Error())
	}

	userName := r.PostFormValue("name")
	userAbout := r.PostFormValue("about")

	file, _, _ := r.FormFile("user_photo")

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	var imageNameA string
	if file != nil {

		// Check weather the string is same as in db
		imageNameA = utils.StoreThisFileInBucket("user_dp_images/", claim.User.UserPhone, file)
		defer file.Close()
	}

	err2 := info.userRepo.UpdateUserProfileDataLogic(userName, userAbout, imageNameA, claim.User.UserPhone)
	if err2 != nil {
		panic(err2.Error())
	}

	// Update data to the database

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserShowPeopleHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	data, err1 := info.userRepo.GetAllUsersLogic(claim.User.UserPhone)
	if err1 != nil {
		panic(err1.Error())
	}

	err := template.Tpl.ExecuteTemplate(w, "user_show_people.html", data)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserNewChatStartedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	message := "+91 " + claim.User.UserPhone + " started a chat with +91 " + target + "."
	data := models.MessageModel{
		Content:     message,
		From:        claim.User.UserPhone,
		To:          target,
		Time:        time.Now().Format("2 Jan 2006 3:04:05 PM"),
		ContentType: logic.TEXT,
		Status:      "ADMIN",
	}

	info.messagesRepo.StorePersonalMessagesLogic(data)

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserCreateGroup(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	// Parse form to get data
	err := r.ParseMultipartForm(10 << 24)
	if err != nil {
		log.Println("File is empty ", err)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	groupName := r.PostFormValue("groupName")
	groupAbout := r.PostFormValue("group-about")

	file, _, _ := r.FormFile("profile_photo")

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	var imageNameA string
	if file != nil {

		// Check weather the string is same as in db
		imageNameA = utils.StoreThisFileInBucket("group_dp_images/", groupName+claim.User.UserPhone, file)
		defer file.Close()
	}

	data := models.GroupModel{
		Image: imageNameA,
		Name:  groupName,
		About: groupAbout,
	}
	claims := &jwtPkg.UserJwtClaim{
		GroupModel: data,
	}

	token := jwtPkg.SignJwtToken(claims)
	expire := time.Now().AddDate(0, 0, 1)
	cookie := &http.Cookie{Name: "userGroupDetails", Value: token, Expires: expire, HttpOnly: true, Path: "/user/dashboard/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/user/dashboard/add-group-members", http.StatusSeeOther)
}

func (info *InformationHelper) UserAddGroupMembers(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	data, err := info.userRepo.GetAllUsersLogic(claim.User.UserPhone)
	if err != nil {
		panic(err.Error())
	}

	err2 := template.Tpl.ExecuteTemplate(w, "add_group_members.html", data)
	if err2 != nil {
		log.Println(err2.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserGroupCreationHandler(w http.ResponseWriter, r *http.Request) {
	type groupMembers struct {
		Data []string `json:"data"`
	}
	var val groupMembers

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &val)
	if err != nil {
		log.Println("ERROR happened -- ", err.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	// Claim to get group data
	gc, err1 := r.Cookie("userGroupDetails")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	groupClaim := jwtPkg.GetValueFromJwt(gc)

	// Claim to get user phone number
	uc, err2 := r.Cookie("userToken")
	if err2 != nil {
		if err2 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	userClaim := jwtPkg.GetValueFromJwt(uc)

	data := models.GroupModel{
		Owner:   userClaim.User.UserPhone,
		Name:    groupClaim.GroupModel.Name,
		About:   groupClaim.GroupModel.About,
		Image:   groupClaim.GroupModel.Image,
		Members: val.Data,
	}

	status, err3 := info.groupRepo.CreateGroupAndInsertDataLogic(data)

	if status {
		fmt.Println("Success - Redirect to dashboard")
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err3.Error())
	}
}

func (info *InformationHelper) UserNewChatSelectedHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(e)
		}
	}()

	type targetID struct {
		Data string `json:"target"`
	}

	var target targetID

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &target)
	if err != nil {
		panic(err.Error())
	}

	uc, err2 := r.Cookie("userToken")
	if err2 != nil {
		if err2 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	userClaim := jwtPkg.GetValueFromJwt(uc)

	// the message status of the message which the target has sent to this user should be marked as read
	err = info.messagesRepo.UpdatePmToRead(target.Data, userClaim.User.UserPhone)
	if err != nil {
		panic(err)
	}

	var (
		uName, uAvtr, uAbout string
		uVal                 []models.MessageModel
	)

	if target.Data != "admin" {
		data, err1 := info.messagesRepo.GetMessageDataLogic(target.Data, userClaim.User.UserPhone)
		if err1 != nil {
			panic(err1.Error())
		}

		uVal = data

		val, err2 := info.userRepo.GetUserDataLogic(target.Data)
		if err2 != nil {
			panic(err2.Error())
		}

		uName = val.UserName
		uAvtr = val.UserAvatarUrl
		uAbout = val.UserAbout
	} else {
		data, err1 := info.messagesRepo.GetMessageDataLogic(target.Data, "all")
		if err1 != nil {
			panic(err1.Error())
		}

		uName = "🔴 ADMIN 🔴"
		uAvtr = ""
		uAbout = "The Administrator"
		uVal = data
	}

	xData := struct {
		Name   string                `json:"name"`
		Avatar string                `json:"avatar"`
		About  string                `json:"about"`
		Val    []models.MessageModel `json:"data"`
	}{
		Name:   uName,
		Avatar: uAvtr,
		About:  uAbout,
		Val:    uVal,
	}

	// If everything goes well...
	s, _ := json.Marshal(xData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (info *InformationHelper) UserGroupChatSelectedHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(e)
		}
	}()

	type targetID struct {
		Data string `json:"target"`
	}

	var target targetID

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &target)
	if err != nil {
		panic(err.Error())
	}

	// Get all messages from the group using group id
	messages, err := info.groupRepo.GetAllGroupMessagesLogic(target.Data)
	if err != nil {
		panic(err)
	}

	// Get all group members details
	uDatas := info.groupRepo.GetAllGroupMembersData(target.Data)

	// Get all group data using the group_id (target)
	// - group_id, group_name, avatar, messages.
	data, err := info.groupRepo.GetGroupDetailsLogic(target.Data)
	if err != nil {
		panic(err)
	}

	uc, err2 := r.Cookie("userToken")
	if err2 != nil {
		if err2 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	userClaim := jwtPkg.GetValueFromJwt(uc)

	isLeft := info.groupRepo.CheckUserLeftTheGroup(userClaim.User.UserPhone, target.Data)

	// make struct and parse data to json format and send
	xData := struct {
		Name         string                     `json:"name"`
		Avatar       string                     `json:"avatar"`
		About        string                     `json:"about"`
		Owner        string                     `json:"owner"`
		Created      string                     `json:"created"`
		TotalMembers int                        `json:"total_members"`
		IsBanned     bool                       `json:"is_banned"`
		IsLeft       bool                       `json:"is_left"`
		BanTime      string                     `json:"ban_time"`
		Members      []models.GroupMembersModel `json:"members_list"`
		Val          []models.MessageModel      `json:"data"`
	}{
		Name:         data.Name,
		Avatar:       data.Image,
		About:        data.About,
		Owner:        data.Owner,
		Created:      data.CreatedDate,
		TotalMembers: data.NoOfMembers,
		IsBanned:     data.IsBanned,
		IsLeft:       isLeft,
		BanTime:      data.BanTime,
		Members:      uDatas,
		Val:          messages,
	}

	// If everything goes well...
	s, _ := json.Marshal(xData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (info *InformationHelper) GroupUnblockHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(e)
		}
	}()

	type targetID struct {
		Data string `json:"target"`
	}

	var target targetID

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &target)
	if err != nil {
		panic(err.Error())
	}

	err = info.userRepo.GroupUnblockLogic(target.Data)
	if err != nil {
		panic(err)
	}

	s, _ := json.Marshal(true)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (info *InformationHelper) UserLeftGroupHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Oh ohh, Theres an error - ", e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()
	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	msg := claim.User.UserPhone + " has left the group."

	err := info.groupRepo.UserLeftTheGroupLogic(target, claim.User.UserPhone, msg)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserKickedOutHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Oh ohh, Theres an error - ", e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()
	vars := mux.Vars(r)
	groupID := vars["group"]
	userID := vars["user"]

	msg := "+91 " + userID + " has been kicked out from the group."

	err := info.groupRepo.UserLeftTheGroupLogic(groupID, userID, msg)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserGroupManagePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	if !info.groupRepo.CheckUserIsAdmin(target, claim.User.UserPhone) {
		panic("User Is not an admin")
	}

	data := info.userRepo.NonGroupMembersLogic(target, claim.User.UserPhone)

	if data == nil {
		panic(errors.New("some error occurred while accessing non group members"))
	}
	xData := struct {
		GroupId string
		Data    []models.UserModel
	}{
		GroupId: target,
		Data:    data,
	}
	err := template.Tpl.ExecuteTemplate(w, "manage_group_members.html", xData)
	if err != nil {
		panic(err)
	}
}

func (info *InformationHelper) UserGroupAddMembersHandler(w http.ResponseWriter, r *http.Request) {
	// Getting slice of user ids
	type groupMembers struct {
		Data []string `json:"data"`
	}
	var val groupMembers

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &val)

	// Getting group id from url
	vars := mux.Vars(r)
	target := vars["target"]

	if err != nil {
		log.Println("ERROR happened -- ", err.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	// redirect to dashboard
	err = info.groupRepo.AddGroupMembers(target, val.Data)
	if err != nil {
		log.Println("ERROR happened -- ", err.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	s, _ := json.Marshal(true)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (info *InformationHelper) UserBlocksHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	err := info.userRepo.UserBlockUserLogic(claim.User.UserPhone, target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserUnblocksHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	fmt.Println("Target = ", target)
	fmt.Println("Yours = ", claim.User.UserPhone)

	err := info.userRepo.UserUnblockUserLogic(claim.User.UserPhone, target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	err := template.Tpl.ExecuteTemplate(w, "about_page.html", http.StatusFound)
	if err != nil {
		panic(err)
	}
}

func (info *InformationHelper) UserDeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	// Use transaction to insert row into table B
	// and then delete a row from table A
	err := info.userRepo.DeleteUserAccountLogic(target)
	if err != nil {
		panic(err)
	}

	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)

	cookie := &http.Cookie{Name: "userToken", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (info *InformationHelper) UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)
	//
	cookie := &http.Cookie{Name: "userToken", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}
