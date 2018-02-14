package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
    "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/dgrijalva/jwt-go"
    "github.com/dgrijalva/jwt-go/request"
    "time"
	"errors"
	"net/http"
	_ "github.com/qor/qor"
	"github.com/qor/admin"
	"github.com/swaggo/gin-swagger"
    "github.com/swaggo/gin-swagger/swaggerFiles"
    _ "github.com/swaggo/gin-swagger/example/docs"
)

// @Description get struct array by ID
// @ID get-struct-array-by-string
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-struct-array-by-string/{some_id} [get]
func main() {
	// DB接続
	db = gormConnect()
	// Admin画面
    launchAdmin()
	// RestAPI設定
	launchRestApi()

}

//------------------------------
// Admin UI
//------------------------------
func launchAdmin(){
    // 管理画面初期化
    Admin := admin.New(&admin.AdminConfig{DB: db})
    // 管理対象のgormテーブルを指定
    Admin.AddResource(&User{})
    // HTTPリクエストマルチプレクサ作成
    mux := http.NewServeMux()
    // 管理画面をマルチプレクサにマウント
    Admin.MountTo("/admin", mux)
    fmt.Println("顧客サービスDB管理画面起動 PORT:9000")
    go http.ListenAndServe(":9000", mux) // 並行処理で起動
}

//------------------------------
// Database
//------------------------------
var db *gorm.DB // グローバル変数としてDBオブジェクトを保持
// エンティティ
type User struct {
    gorm.Model
    Email     string `json:"user_id"`
    Password  string `json:"password"`
}

// DB接続
func gormConnect() *gorm.DB {
    DBMS     := "mysql"
    USER     := "root"
    PASS     := "mysql"
    PROTOCOL := "tcp(user-mysql:3306)"
    DBNAME   := "micro_user"
    // DB接続
    CONNECT := USER+":"+PASS+"@"+PROTOCOL+"/"+DBNAME+"?charset=utf8&parseTime=true"
    db,err := gorm.Open(DBMS, CONNECT)
    if err != nil {
        panic(err.Error())
    } else {
        fmt.Println("DB接続成功")
	}
	// テーブル作成
    if !db.HasTable(&User{}) {
        db.CreateTable(&User{})
    }
    return db
}

//------------------------------
// REST API
//------------------------------
// REST API起動
func launchRestApi(){
    r := gin.Default()
    // CORS設定
    config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    config.AllowHeaders = []string{"Authorization"}
    r.Use(cors.New(config))

	r.GET("/", IndexHandler) // APIトップ
	r.POST("/login", LoginHandler) // ログイン用APIエンドポイントハンドラー
	r.POST("/user", CreateUserHandler) // ユーザ登録用APIエンドポイントハンドラー
	r.GET("/me", CurrentUserHandler) // 自ユーザ情報取得エンドポイントハンドラー
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.Run(":3000")
    fmt.Println("顧客サービス起動完了")
}

// サービスルートエンドポイント
func IndexHandler(c *gin.Context) {
    c.JSON(200, gin.H{"message": "顧客サービスへようこそ!!"})
}

// エラーレスポンスオブジェクト定義
type ErrorResponse struct{
    ErrorCode int `json: error_code`
    Message string `json error_message`
}


// @Description get struct array by ID
// @ID get-struct-array-by-string
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-struct-array-by-string/{some_id} [get]
func LoginHandler(c *gin.Context) {
    // リクエストパラメータ取得
    email := c.PostForm("email") 
    password := c.PostForm("password")
    // ログイン認証
    user := User{}
    if db.Where(&User{Email: email, Password: password}).First(&user).Error == nil{
		// JWTトークン生成してレスポンスとして返却
		token,err := generateJwtToken(&user)
        if err != nil{
            c.JSON(400,ErrorResponse{1,"JWT生成に失敗しました"})
        } else {
            c.JSON(200, gin.H{"token": token})
        }
    } else {
        c.JSON(400,ErrorResponse{1,"ログインに失敗しました"})
    }
}

// ユーザ登録用エンドポイント
func CreateUserHandler(c *gin.Context) {
    // リクエストパラメータ取得
    email := c.PostForm("email")
    password := c.PostForm("password")
    // DBにユーザを登録
    user := User{ Email: email, Password:password }
    db.NewRecord(user)
    db.Create(&user)
    db.Save(&user)
    c.JSON(200,user)
}

//------------------------------
// 認証
//------------------------------
// ユーザ情報取得
func CurrentUserHandler(c *gin.Context){
    // 署名の検証
    token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
        b := []byte(secretKey)
        return b, nil
    })
    if err == nil {
        claims := token.Claims.(jwt.MapClaims)
        c.JSON(200, gin.H{"user_id":claims["user_id"],"email":claims["email"]})
    } else {
        c.JSON(400,ErrorResponse{1,"トークン検証に失敗しました" + err.Error()})
    }
}
//------------------------------
// JWTトークン生成
//------------------------------
var secretKey = "75c92a074c341e9964329c0550c2673730ed8479c885c43122c90a2843177d5ef21cb50cfadcccb20aeb730487c11e09ee4dbbb02387242ef264e74cbee97213"
// トークン生成
func generateJwtToken(user *User) (string, error){
    // アルゴリズム指定
    token := jwt.New(jwt.GetSigningMethod("HS256"))
    // トークン生成
    token.Claims = jwt.MapClaims{
        "user_id": user.ID,
        "email": user.Email,
        "exp":  time.Now().Add(time.Hour * 24).Unix(),
    }
    // 署名付与
    tokenString, err := token.SignedString([]byte(secretKey))
    if err == nil {
        return tokenString,nil
    } else {
        return "",errors.New("トークン生成エラー")
    }
}