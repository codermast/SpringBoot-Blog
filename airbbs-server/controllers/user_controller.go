package controllers

import (
	"codermast.com/airbbs/config"
	"codermast.com/airbbs/constant"
	"codermast.com/airbbs/daos"
	"codermast.com/airbbs/models/po"
	"codermast.com/airbbs/models/ro"
	"codermast.com/airbbs/models/vo"
	"codermast.com/airbbs/services"
	"codermast.com/airbbs/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"log"
	"net/http"
	"path/filepath"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// GetAllUsers 获取所有用户 Get /user/all
func (uc *UserController) GetAllUsers(c *gin.Context) {
	userVos := services.GetAllUsers()

	// 处理获取所有用户的逻辑
	c.JSON(http.StatusOK, utils.Success("查询成功", userVos))
}

// GetUserByID 根据 ID 获取指定用户 Get /user/:uid
func (uc *UserController) GetUserByID(c *gin.Context) {
	// 获取路径参数
	userID := c.Param("uid")

	userVo, err := services.GetUserByID(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}
	// 处理根据用户ID获取用户信息的逻辑
	c.JSON(http.StatusOK, utils.SuccessData(userVo))
}

// CreateUser 创建用户 POST /user/register
func (uc *UserController) CreateUser(c *gin.Context) {

	var userRegisterRequest ro.UserRegisterRequest

	// 用户解析
	if err := c.BindJSON(&userRegisterRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 获取验证码
	code := userRegisterRequest.Code

	if code == "" {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("验证码为空！")))
		return
	}

	// 构建 User 对象
	var user po.User

	_ = copier.Copy(&user, &userRegisterRequest)

	// 默认创建的用户为非管理员
	user.Admin = false

	// 账号或密码为空
	if (user.Username == "") || (user.Password == "") {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("username or password is empty")))
		return
	}

	// 校验验证码
	redisCode, err := utils.Get(fmt.Sprintf("%s:%s", "register", user.Mail))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("请先获取验证码")))
		return
	}

	if redisCode != code {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("验证码错误！")))
		return
	}

	// 保存用户信息
	err = services.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusConflict, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 此时验证码正确且用户注册成功，则通过校验，删除 Redis 中验证码，减少资源占用
	err = utils.Del(userRegisterRequest.Mail)
	if err != nil {
		log.Println("验证码删除失败")
	}

	var userVO vo.UserVO

	_ = copier.Copy(&userVO, &user)

	if err != nil {
		log.Println("结构体赋值异常！")
	}

	// 处理根据用户ID获取用户信息的逻辑
	c.JSON(http.StatusOK, utils.Success("注册成功！", userVO))
}

// UpdateUser 更新指定 userID 的用户信息 PUT /user/:uid
func (uc *UserController) UpdateUser(c *gin.Context) {

	userId := c.Param("uid")

	var userRo ro.UserUpdateInfoRequest

	// 用户解析
	if err := c.BindJSON(&userRo); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error("数据格式错误！"))
		return
	}

	// 判断 JWT 中 UserID 和 UserVo 中是否匹配
	if c.GetString(constant.USERID) != userRo.ID || userId != c.GetString(constant.USERID) {
		c.JSON(http.StatusBadRequest, utils.Error("用户ID不匹配！"))
		return
	}

	err := services.UpdateUser(&userRo)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 处理根据用户ID获取用户信息的逻辑
	c.JSON(http.StatusOK, utils.SuccessMsg("更新成功！"))
}

// DeleteUser 删除指定 userID 的用户 DELETE /user/:uid
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("uid")

	err := services.DeleteUserByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 处理根据用户ID获取用户信息的逻辑
	c.JSON(http.StatusOK, utils.SuccessMsg("删除成功！"))
}

// UserLogin 用户登录 POST /user/login
func (uc *UserController) UserLogin(c *gin.Context) {
	var userLoginRo ro.UserLoginRequest

	// 用户解析
	if err := c.BindJSON(&userLoginRo); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error("用户数据解析失败！"))
		return
	}

	if userLoginRo.Username == "" || userLoginRo.Password == "" {
		c.JSON(http.StatusBadRequest, utils.Error("用户名或账号密码为空！"))
		return
	}

	var user po.User

	_ = copier.Copy(&user, &userLoginRo)

	err := services.UserLogin(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 登录成功则设置 JWT token
	tokenString, err := utils.GetJwtToken(user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error("Failed to generate token"))
		return
	}

	response := &vo.UserLoginVo{
		Token:  tokenString,
		UserId: user.ID,
	}

	// 将 Token 返回给客户端
	c.JSON(http.StatusOK, utils.Success("登录成功", response))
}

// ResetUserPassword 重置用户密码 /users/password/reset
func (uc *UserController) ResetUserPassword(c *gin.Context) {
	var userResetPasswordRequest ro.UserResetPasswordRequest

	// 用户解析
	if err := c.BindJSON(&userResetPasswordRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error("数据格式错误！"))
		return
	}

	account := userResetPasswordRequest.Account
	password := userResetPasswordRequest.Password
	code := userResetPasswordRequest.Code

	// 账户为空返回
	if account == "" {
		c.JSON(http.StatusConflict, utils.Error("账户不能为空！"))
		return
	}

	// 根据账户获取邮箱账号
	mail, err := daos.GetMailByAccount(account)

	if err != nil {
		c.JSON(http.StatusConflict, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	// 获取验证码
	redisCode, err := utils.Get(fmt.Sprintf("%s:%s", "reset", mail))

	if err != nil {
		c.JSON(http.StatusConflict, utils.Error("请先发送验证码！"))
		return
	}

	// 判断验证码是否正确
	if redisCode != code {
		c.JSON(http.StatusConflict, utils.Error("验证码错误！"))
		return
	}

	// 此时验证码匹配成功，开始更新密码

	err = services.ResetUserPassword(account, password)

	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessMsg("重置成功！"))
}

// UploadUserPhoto 更新用户头像 /users/photo
func (uc *UserController) UploadUserPhoto(c *gin.Context) {
	// 获取表单文件
	file, err := c.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	// 定义保存文件的目录和文件名
	saveDir := "./uploads"
	savePath := filepath.Join(saveDir, file.Filename)

	// 保存文件到指定目录
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	url := "//" + config.GetServerConfig().Host + ":" + config.GetServerConfig().Port + "/" + savePath

	// 用户 ID
	userID := c.GetString(constant.USERID)

	err = services.UpdateUserPhoto(url, userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(fmt.Sprintf("%v", err)))
		return
	}

	c.JSON(http.StatusOK, utils.Success("更新成功！", url))
}
