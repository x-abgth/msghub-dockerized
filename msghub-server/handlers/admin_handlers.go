package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	jwtPkg "github.com/x-abgth/msghub-dockerized/msghub-server/utils/jwt"

	"github.com/gorilla/mux"
	"github.com/x-abgth/msghub-dockerized/msghub-server/logic"
	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
	"github.com/x-abgth/msghub-dockerized/msghub-server/template"
	"github.com/x-abgth/msghub-dockerized/msghub-server/utils"
)

type AdminHandlerStruct struct {
	logics    logic.AdminDb
	msgLogics logic.MessageDb
	err       error
}

func (admin *AdminHandlerStruct) AdminLoginPageHandler(w http.ResponseWriter, r *http.Request) {

	err1 := admin.logics.MigrateAdminDb()
	if err1 != nil {
		log.Println("Error creating user table : ", err1.Error())
		os.Exit(1)
	}

	c, err1 := r.Cookie("adminToken")
	if err1 != nil {
		type adminLoginData struct {
			ErrStr string
		}

		var data adminLoginData
		if admin.err != nil {
			data = adminLoginData{
				ErrStr: admin.err.Error(),
			}
		}

		err := template.Tpl.ExecuteTemplate(w, "admin_login.html", data)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	} else if jwtPkg.GetValueFromAdminJwt(c).IsAuthenticated {
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	} else {
		panic("An unknown error occured while getting the cookie!")
	}

}

func (admin *AdminHandlerStruct) AdminAuthenticateHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	name := r.PostFormValue("signinName")
	pass := r.PostFormValue("signinPass")

	alert := admin.logics.AdminLoginLogic(name, pass)
	admin.err = alert

	if alert == nil {
		admin.err = nil

		// remove user cookie
		userCookie := &http.Cookie{Name: "userToken", MaxAge: -1, HttpOnly: true, Path: "/"}
		http.SetCookie(w, userCookie)

		// set admin cookie
		claims := &jwtPkg.AdminJwtClaim{
			AdminName:       name,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignAdminJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		adminCookie := &http.Cookie{Name: "adminToken", Value: token, Expires: expire, HttpOnly: true, Path: "/admin/"}
		http.SetCookie(w, adminCookie)

		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin/login-page", http.StatusSeeOther)
	}
}

func (admin *AdminHandlerStruct) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/login-page", http.StatusSeeOther)
		}
	}()

	// Get admin name
	cookie, err1 := r.Cookie("adminToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromAdminJwt(cookie)

	// Get admin table content
	a, err := admin.logics.GetAllAdminsData(claim.AdminName)
	if err != nil {
		panic(err.Error())
	}

	// Get Users table content
	b, err := admin.logics.GetUsersData()
	if err != nil {
		panic(err)
	}

	// Get Deleted Users table content
	c, err := admin.logics.GetDelUsersData()
	if err != nil {
		panic(err)
	}

	// Get Groups table content
	d, err := admin.logics.GetGroupsData()
	if err != nil {
		panic(err)
	}

	// Set data
	data := models.AdminDashboardModel{
		AdminName:             claim.AdminName,
		AdminTbContent:        a,
		UsersTbContent:        b,
		DeletedUsersTbContent: c,
		GroupTbContent:        d,
	}

	err = template.Tpl.ExecuteTemplate(w, "admin_dashboard.html", data)
	if err != nil {
		panic(err)
	}
}

func (admin *AdminHandlerStruct) AdminBlocksUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["id"]
	condition := vars["condition"]

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	var t string

	switch condition {
	case "day":
		t = time.Now().Add(time.Hour * 24).Format("2-01-2006 3:04:05 PM")
		fmt.Println(t)
	case "week":
		t = time.Now().Add(time.Hour * 168).Format("2-01-2006 3:04:05 PM")
		fmt.Println(t)
	case "month":
		t = time.Now().Add(time.Hour * 720).Format("2-01-2006 3:04:05 PM")
		fmt.Println(t)
	case "permanent":
		t = "permanent"
	default:
		log.Println("Sorry, wrong choice.")
		panic("wrong choice")
	}

	// change block value to true and update block duration
	err := admin.logics.BlockThisUserLogic(uid, t)
	if err != nil {
		panic(err)
	}

	// clear cookie and check user block while login
	userCookie := &http.Cookie{Name: "userToken", MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, userCookie)

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) AdminUnBlocksUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["id"]

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	err := admin.logics.UnblockUserLogic(uid)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) AdminBlocksGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gid := vars["id"]
	condition := vars["condition"]

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	var t string
	switch condition {
	case "day":
		t = time.Now().Add(time.Hour * 24).Format("Jan 2, 2006 03:04:05 PM")
		fmt.Println(t)
	case "week":
		t = time.Now().Add(time.Hour * 168).Format("Jan 2, 2006 03:04:05 PM")
		fmt.Println(t)
	case "month":
		t = time.Now().Add(time.Hour * 720).Format("Jan 2, 2006 03:04:05 PM")
		fmt.Println(t)
	case "permanent":
		t = "permanent"
	default:
		log.Println("Sorry, wrong choice.")
		panic("wrong choice")
	}

	err := admin.logics.BlockThisGroupLogic(gid, t)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) AdminUnBlockGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	err := admin.logics.AdminUnBlockGroupHandler(id)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) AdminBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	msg := r.PostFormValue("message")

	data := models.MessageModel{
		Content:     msg,
		From:        "admin",
		To:          "all",
		Time:        time.Now().Format("2 Jan 2006 3:04:05 PM"),
		Status:      "SENT",
		ContentType: logic.TEXT,
	}

	admin.msgLogics.StorePersonalMessagesLogic(data)

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) NewAdminPageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	if err := template.Tpl.ExecuteTemplate(w, "add_new_admin.html", nil); err != nil {
		panic(err)
	}
}

func (admin *AdminHandlerStruct) NewAdminHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}
	}()

	// Get admin name and password

	r.ParseForm()

	name := r.PostFormValue("adminName")
	pass := r.PostFormValue("adminPass1")

	// hash password
	encryptedFormPassword, err := utils.HashEncrypt(pass)
	if err != nil {
		panic(err)
	}

	// store password
	err = admin.logics.InsertAdminLogic(name, encryptedFormPassword)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

func (admin *AdminHandlerStruct) AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	adminCookie := &http.Cookie{Name: "adminToken", MaxAge: -1, HttpOnly: true, Path: "/admin/"}
	http.SetCookie(w, adminCookie)

	http.Redirect(w, r, "/admin/login-page", http.StatusFound)
}
