## MSGHUB - CHAT APPLICATION
<div style="display: flex;"> 
  <img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="golang" />&nbsp;&nbsp;&nbsp;
  <img alt="GitHub last commit (branch)" src="https://img.shields.io/github/last-commit/x-abgth/msghub/main">&nbsp;&nbsp;&nbsp;
  <img src="https://img.shields.io/github/stars/x-abgth/msghub.svg" alt="golang" />
</div>
A messaging application built using <strong>golang</strong> and<strong> bootstrap</strong>, one can group message and personal message.<br>
The front-end section of the application consist of HTML, CSS, JS & Bootstrap and the backend is completely built on Golang (GO). And to access websocket the developer has used the help of gorilla/websocket.<br><br>

> Any UI based contributions are highly encouraged.

## Features 
- Personal Messaging
- Group Messaging
- Story / Status viewing and adding
- Group Creation
<br>

> The app do have limitations because the front-end is just html, css and js.
> Therefore Live reload may not be possible like react or other front-end frameworks.

## Screenshots
<img src="ui_scrnshots/1.png" alt="Home Page" width="800"/><br><br>
<img src="ui_scrnshots/2.png" alt="Home Page" width="800"/><br><br>
<img src="ui_scrnshots/3.png" alt="Home Page" width="800"/><br><br>
<img src="ui_scrnshots/4.png" alt="Home Page" width="800"/><br><br>

## 💻 Test application on your machine
- Create ```.env``` file in the ```msghub-server``` directory which should include -
  - TWILIO_SID
  - TWILIO_TOKEN
  - TWILIO_SERVICE
  - DB_HOST
  - DB_PORT
  - DB_USER
  - DB_PASS
  - DB_NAME
  - DB_SSLMODE
  - AWS_S3_REGION
  - AWS_S3_BUCKET
  - JWT_KEY
- Then open terminal from root directory of this application and run :
```
cd msghub-server
```
```
go mod tidy
```
```
go run main.go
```
- Then open ```http://localhost:9000/``` in your browser.

> The websocket might not work when you running this application locally because websocket runs on <strong>wss</strong> protocol.

## ❤ Conclusion
🌟 Star this repo & follow for more 😊

<a href="https://www.buymeacoffee.com/abgth" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>