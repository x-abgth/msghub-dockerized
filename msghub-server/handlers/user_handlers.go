package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/template"
	"github.com/x-abgth/msghub-dockerized/msghub-server/utils"
	jwtPkg "github.com/x-abgth/msghub-dockerized/msghub-server/utils/jwt"

	"github.com/x-abgth/msghub-dockerized/msghub-server/logic"

	"github.com/gorilla/mux"

	"net/http"
)

// This might helps to pass error strings from one route to other
type UserHandler struct {
	migrationService logic.MigrationLogic
	userService      logic.UserLogic
}

func NewUserHandler(migrationServ logic.MigrationLogic, userServ logic.UserLogic) *UserHandler {
	return &UserHandler{migrationService: migrationServ, userService: userServ}
}

func (u *UserHandler) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := u.migrationService.MigrateUserTable()
	if err != nil {
		log.Fatal("Error creating user table : ", err.Error())
	}

	err = u.migrationService.MigrateDeletedUserTable()
	if err != nil {
		log.Fatal("Error creating deleted user table : ", err.Error())
	}

	// Setting jwt token
	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}
	token := jwtPkg.SignJwtToken(claims)
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    token,
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	// Checking login error message
	errCookie, err := r.Cookie("login_error")
	if err != nil {
		errCookie = &http.Cookie{Name: "login_error", Value: ""}
	}

	alm := struct {
		ErrorStr string
	}{
		ErrorStr: errCookie.Value,
	}

	err = template.Tpl.ExecuteTemplate(w, "index.html", alm)
	if err != nil {
		fmt.Println("Error : ", err.Error())
	}
}

func (u *UserHandler) UserLoginCredentialsHandler(w http.ResponseWriter, r *http.Request) {

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

	err := u.userService.UserLoginLogic(ph, pass)
	if err == nil {
		// assigning JWT tokens
		claims := &jwtPkg.UserJwtClaim{
			UserID:          ph,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "user_token", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "login_error",
			Value:    err.Error(),
			MaxAge:   60, // 60 secs
			HttpOnly: true,
			Path:     "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// This handler displays the page to enter the phone number
func (u *UserHandler) UserLoginWithOtpPageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	loginErr, err := r.Cookie("login_error")
	if err != nil {
		loginErr = &http.Cookie{Name: "login_error", Value: ""}
	}

	data := struct {
		ErrorStr string
	}{
		ErrorStr: loginErr.Value,
	}
	err = template.Tpl.ExecuteTemplate(w, "login_with_otp.html", data)
	if err != nil {
		panic(err)
	}
}

// This handler process the phone number given and check weather is valid or not
func (u *UserHandler) UserVerifyLoginPhoneHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	r.ParseForm()
	ph := r.PostFormValue("phone")

	var result bool
	if len(ph) == 10 {
		err := u.userService.UserDuplicationStatsAndSendOtpLogic(ph)
		if err != nil {
			log.Println(err)
			result = false
		} else {
			result = true
		}
	}

	// Create a JSON response with the result
	response := struct {
		Result bool `json:"result"`
	}{
		Result: result,
	}

	// Encode the response as JSON and write it to the response writer
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *UserHandler) UserVerifyLoginOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	otp := r.PostFormValue("loginOtp")
	ph := r.PostFormValue("phone")

	err := u.userService.UserValidateOtpLogic(ph, otp)

	if err == nil {
		claims := &jwtPkg.UserJwtClaim{
			UserID:          ph,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "user_token", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "login_error",
			Value:    "Entered OTP is invalid.",
			MaxAge:   60, // 60 secs
			HttpOnly: true,
			Path:     "/login/phone",
		})

		http.Redirect(w, r, "/login/phone/validation", http.StatusFound)
	}
}

func (u *UserHandler) UserVerifyRegisterPhoneHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	ph := r.PostFormValue("phone")

	var result bool
	// This only sends the otp for registration, won't update anything on db
	err := u.userService.UserRegisterPhoneValidationLogic(ph)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "login_error",
			Value:    err.Error(),
			MaxAge:   60, // 60 secs
			HttpOnly: true,
			Path:     "/",
		})
		result = false
	} else {
		result = true
	}

	result = true

	// Create a JSON response with the result
	response := struct {
		Result bool `json:"result"`
	}{
		Result: result,
	}

	// Encode the response as JSON and write it to the response writer
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *UserHandler) UserVerifyRegisterOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.PostFormValue("signupName")
	phone := r.PostFormValue("signupPh")
	pass := r.PostFormValue("signupPass")
	otp := r.PostFormValue("otp")

	if ok := u.userService.UserRegisterLogic(otp, name, phone, pass); ok {

		claims := &jwtPkg.UserJwtClaim{
			UserID:          phone,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "user_token", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {

		http.SetCookie(w, &http.Cookie{
			Name:     "login_error",
			Value:    "Account registration failed. Invalid OTP.",
			MaxAge:   60, // 60 secs
			HttpOnly: true,
			Path:     "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (u *UserHandler) UserDashboardHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			log.Println(e)
			http.SetCookie(w, &http.Cookie{
				Name:     "login_error",
				Value:    "Sorry, An unknown error occured",
				MaxAge:   60, // 60 secs
				HttpOnly: true,
				Path:     "/",
			})
			cookie := &http.Cookie{Name: "user_token", MaxAge: -1, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	// Creates table for user stories
	storyErr := u.migrationService.MigrateStoriesTable()
	if storyErr != nil {
		panic(storyErr.Error())
	}

	// Creates table for user's personal messages
	migrateErr := u.migrationService.MigrateMessageTable()
	if migrateErr != nil {
		panic(migrateErr.Error())
	}

	// Group migration statements
	groupMigrationError := u.migrationService.MigrateGroupTable()
	if groupMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupMigrationError.Error())
	}

	groupUserMigrationError := u.migrationService.MigrateUserGroupRelationTable()
	if groupUserMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupUserMigrationError.Error())
	}

	groupMessageMigrationError := u.migrationService.MigrateGroupMessageTable()
	if groupMessageMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupMessageMigrationError.Error())
	}

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	// Update every sent status of the user to delivered, when the user gets online
	err := u.userService.UpdatePmToDelivered(userId)
	if err != nil {
		panic(err)
	}

	data, err := u.userService.GetDataForDashboardLogic(userId)
	if err != nil {
		log.Println("Error getting dashboard logic")
		panic(err.Error())
	}

	err2 := template.Tpl.ExecuteTemplate(w, "user_dashboard.html", data)
	if err2 != nil {
		log.Println(err2)
		panic("Not yet in dashboard page")
	}
}

func (u *UserHandler) UserAddStoryHandler(w http.ResponseWriter, r *http.Request) {
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
	err := u.userService.AddNewStoryLogic(target, imageNameA)
	if err != nil {
		panic(err)
	}

	// redirect to the dashboard
	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserStorySeenHandler(w http.ResponseWriter, r *http.Request) {
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

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	// Add story viewers
	err := u.userService.StorySeenLogic(userId, target)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(nil)
}

func (u *UserHandler) UserDeleteStoryHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	if target != userId {
		panic("Invalid access to delete story")
	}

	// Delete story
	err := u.userService.DeleteUserStoryLogic(target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			log.Println(e)
			cookie := &http.Cookie{Name: "user_token", MaxAge: -1, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}()

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	// Take values from the database
	userInfo, err2 := u.userService.GetUserDataLogic(userId)
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

func (u *UserHandler) UserProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {

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

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	var imageNameA string
	if file != nil {

		// Check weather the string is same as in db
		imageNameA = utils.StoreThisFileInBucket("user_dp_images/", userId, file)
		defer file.Close()
	}

	err2 := u.userService.UpdateUserProfileDataLogic(userName, userAbout, imageNameA, userId)
	if err2 != nil {
		panic(err2.Error())
	}

	// Update data to the database

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserShowPeopleHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	data, err1 := u.userService.GetAllUsersLogic(userId)
	if err1 != nil {
		panic(err1.Error())
	}

	err := template.Tpl.ExecuteTemplate(w, "user_show_people.html", data)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (u *UserHandler) UserNewChatStartedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	message := "+91 " + userId + " started a chat with +91 " + target + "."
	data := models.MessageModel{
		Content:     message,
		From:        userId,
		To:          target,
		Time:        time.Now().Format("2 Jan 2006 3:04:05 PM"),
		ContentType: logic.TEXT,
		Status:      "ADMIN",
	}

	u.userService.StorePersonalMessagesLogic(data)

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserCreateGroup(w http.ResponseWriter, r *http.Request) {

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

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	var imageNameA string
	if file != nil {

		// Check weather the string is same as in db
		imageNameA = utils.StoreThisFileInBucket("group_dp_images/", groupName+userId, file)
		defer file.Close()
	}

	data := models.GroupModel{
		Image: imageNameA,
		Name:  groupName,
		About: groupAbout,
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		panic("json encoding failed")
	}

	encoded64 := base64.StdEncoding.EncodeToString(encoded)

	expire := time.Now().AddDate(0, 0, 1)
	cookie := &http.Cookie{Name: "userGroupDetails", Value: encoded64, Expires: expire, HttpOnly: true, Path: "/user/dashboard/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/user/dashboard/add-group-members", http.StatusSeeOther)
}

func (u *UserHandler) UserAddGroupMembers(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	data, err := u.userService.GetAllUsersLogic(userId)
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

func (u *UserHandler) UserGroupCreationHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	type groupMembers struct {
		Data []string `json:"data"`
	}
	var val groupMembers

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &val)
	if err != nil {
		panic(err)
	}

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	// Claim to get group data
	gc, err1 := r.Cookie("userGroupDetails")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	encoded, err := base64.StdEncoding.DecodeString(gc.Value)
	if err != nil {
		panic("Decoding failed")
	}

	var group models.GroupModel
	err = json.Unmarshal(encoded, &group)
	if err != nil {
		panic("unmarshaling failed")
	}

	data := models.GroupModel{
		Owner:   userId,
		Name:    group.Name,
		About:   group.About,
		Image:   group.Image,
		Members: val.Data,
	}

	status, err3 := u.userService.CreateGroupAndInsertDataLogic(data)

	if status {
		fmt.Println("Success - Redirect to dashboard")
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err3.Error())
	}
}

func (u *UserHandler) UserNewChatSelectedHandler(w http.ResponseWriter, r *http.Request) {
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

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &target)
	if err != nil {
		panic(err.Error())
	}

	// the message status of the message which the target has sent to this user should be marked as read
	err = u.userService.UpdatePmToRead(target.Data, userId)
	if err != nil {
		panic(err)
	}

	var (
		uName, uAvtr, uAbout string
		uVal                 []models.MessageModel
	)

	if target.Data != "admin" {
		data, err1 := u.userService.GetMessageDataLogic(target.Data, userId)
		if err1 != nil {
			panic(err1.Error())
		}

		uVal = data

		val, err2 := u.userService.GetUserDataLogic(target.Data)
		if err2 != nil {
			panic(err2.Error())
		}

		uName = val.UserName
		uAvtr = val.UserAvatarUrl
		uAbout = val.UserAbout
	} else {
		data, err1 := u.userService.GetMessageDataLogic(target.Data, "all")
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

func (u *UserHandler) UserGroupChatSelectedHandler(w http.ResponseWriter, r *http.Request) {
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

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &target)
	if err != nil {
		panic(err.Error())
	}

	// Get all messages from the group using group id
	messages, err := u.userService.GetAllGroupMessagesLogic(target.Data)
	if err != nil {
		panic(err)
	}

	// Get all group members details
	uDatas := u.userService.GetAllGroupMembersData(target.Data)

	// Get all group data using the group_id (target)
	// - group_id, group_name, avatar, messages.
	data, err := u.userService.GetGroupDetailsLogic(target.Data)
	if err != nil {
		panic(err)
	}

	isLeft := u.userService.CheckUserLeftTheGroup(userId, target.Data)

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

func (u *UserHandler) GroupUnblockHandler(w http.ResponseWriter, r *http.Request) {
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

	err = u.userService.GroupUnblockLogic(target.Data)
	if err != nil {
		panic(err)
	}

	s, _ := json.Marshal(true)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (u *UserHandler) UserLeftGroupHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Oh ohh, Theres an error - ", e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()
	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	msg := userId + " has left the group."

	err := u.userService.UserLeftTheGroupLogic(target, userId, msg)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserKickedOutHandler(w http.ResponseWriter, r *http.Request) {
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

	err := u.userService.UserLeftTheGroupLogic(groupID, userID, msg)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserGroupManagePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	if !u.userService.CheckUserIsAdmin(target, userId) {
		panic("User Is not an admin")
	}

	data := u.userService.NonGroupMembersLogic(target, userId)

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

func (u *UserHandler) UserGroupAddMembersHandler(w http.ResponseWriter, r *http.Request) {
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
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	// redirect to dashboard
	err = u.userService.AddGroupMembers(target, val.Data)
	if err != nil {
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	s, _ := json.Marshal(true)
	w.Header().Set("Content-Type", "application/json")
	w.Write(s)
}

func (u *UserHandler) UserBlocksHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	err := u.userService.UserBlockUserLogic(userId, target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) UserUnblocksHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	vars := mux.Vars(r)
	target := vars["target"]

	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		panic("user id is not found")
	}

	fmt.Println("Target = ", target)
	fmt.Println("Yours = ", userId)

	err := u.userService.UserUnblockUserLogic(userId, target)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (u *UserHandler) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
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

func (u *UserHandler) UserDeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
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
	err := u.userService.DeleteUserAccountLogic(target)
	if err != nil {
		panic(err)
	}

	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)

	cookie := &http.Cookie{Name: "user_token", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (u *UserHandler) UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)
	//
	cookie := &http.Cookie{Name: "user_token", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}
