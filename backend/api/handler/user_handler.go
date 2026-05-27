package handler

import (
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/api/validate"
	"awesomeProject4/backend/repository/dao"
	"awesomeProject4/pkg"
	"awesomeProject4/pkg/zapx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) registerUser(c *gin.Context) {
	user := validate.UserValidate(c)
	if user == nil {
		return
	}

	passwordHash, err := pkg.Encrypt(user.Password)
	if err != nil {
		h.l.Error("encrypt user password failed", zapx.Error(err), zapx.String("username", user.Username))
		utils.InternalServerError(c, err.Error())
		return
	}

	model := &dao.User{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: passwordHash,
	}

	if err := h.userService.Register(c.Request.Context(), model); err != nil {
		h.l.Error("register user failed", zapx.Error(err), zapx.String("username", user.Username), zapx.String("email", user.Email))
		utils.InternalServerError(c, err.Error())
		return
	}

	token, err := utils.GenerateToken(int64(model.ID), model.Username)
	if err != nil {
		h.l.Error("generate register token failed", zapx.Error(err), zapx.String("username", model.Username))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "user registered", gin.H{
		"username": user.Username,
		"email":    user.Email,
		"token":    token,
	})
}

func (h *Handler) loginUser(c *gin.Context) {
	user := validate.UserValidate(c)
	if user == nil {
		return
	}
	data, err := h.userService.Login(c.Request.Context(), user.Email)
	if err != nil {
		h.l.Error("find user for login failed", zapx.Error(err), zapx.String("email", user.Email))
		utils.Unauthorized(c, err.Error())
		return
	}
	password, err := pkg.Decrypt(data.PasswordHash)
	if err != nil {
		h.l.Error("decrypt user password failed", zapx.Error(err), zapx.String("username", data.Username), zapx.String("email", data.Email))
		utils.InternalServerError(c, err.Error())
		return
	}
	if password != user.Password {
		utils.Unauthorized(c, "email or password is incorrect")
		return
	}

	token, err := utils.GenerateToken(int64(data.ID), data.Username)
	if err != nil {
		h.l.Error("generate login token failed", zapx.Error(err), zapx.String("username", data.Username), zapx.String("email", data.Email))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "login success", gin.H{
		"id":       data.ID,
		"username": data.Username,
		"email":    data.Email,
		"token":    token,
	})
}

func (h *Handler) logoutUser(c *gin.Context) {
	utils.NotImplemented(c, "user logout route reserved")
}

func (h *Handler) CreateUserModel(c *gin.Context) {
	userModel := validate.UserModelValidate(c)
	if userModel == nil {
		return
	}

	apiKeyEncrypted, err := pkg.Encrypt(userModel.APIKey)
	if err != nil {
		h.l.Error("encrypt user model api key failed", zapx.Error(err), zapx.String("model_name", userModel.ModelName))
		utils.InternalServerError(c, err.Error())
		return
	}

	model := &dao.UserModel{
		UserID:          0,
		Name:            userModel.Name,
		ModelName:       userModel.ModelName,
		Protocol:        "openai",
		BaseURL:         userModel.BaseURL,
		APIKeyEncrypted: apiKeyEncrypted,
		ProviderName:    userModel.ProviderName,
		IsDefault:       int(userModel.IsDefault),
	}

	if err := h.userModelService.Create(c.Request.Context(), model); err != nil {
		h.l.Error("create user model failed", zapx.Error(err), zapx.String("model_name", userModel.ModelName))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "user model created", model)
}

func (h *Handler) UpdateUserModel(c *gin.Context) {
	id, ok := validate.ParseUintIDParam(c, "id")
	if !ok {
		return
	}

	userModel := validate.UserModelValidate(c)
	if userModel == nil {
		return
	}

	apiKeyEncrypted, err := pkg.Encrypt(userModel.APIKey)
	if err != nil {
		h.l.Error("encrypt user model api key failed", zapx.Error(err), zapx.String("model_name", userModel.ModelName))
		utils.InternalServerError(c, err.Error())
		return
	}

	model := &dao.UserModel{
		ID:              id,
		Name:            userModel.Name,
		ModelName:       userModel.ModelName,
		Protocol:        "openai",
		BaseURL:         userModel.BaseURL,
		APIKeyEncrypted: apiKeyEncrypted,
		ProviderName:    userModel.ProviderName,
		IsDefault:       int(userModel.IsDefault),
	}

	if err := h.userModelService.Update(c.Request.Context(), model); err != nil {
		h.l.Error("update user model failed", zapx.Error(err), zapx.String("model_name", userModel.ModelName))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "user model updated", model)
}

func (h *Handler) DeleteUserModel(c *gin.Context) {
	id, ok := validate.ParseUintIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.userModelService.Delete(c.Request.Context(), id); err != nil {
		h.l.Error("delete user model failed", zapx.Error(err), zapx.Int64("id", int64(id)))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "user model deleted", gin.H{
		"id": id,
	})
}
func (h *Handler) GetUserModel(c *gin.Context) {
	id, ok := validate.ParseUintIDParam(c, "id")
	if !ok {
		return
	}

	model, err := h.userModelService.Get(c.Request.Context(), id)
	if err != nil {
		h.l.Error("get user model failed", zapx.Error(err), zapx.Int64("id", int64(id)))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, model)
}
func (h *Handler) listUserModel(c *gin.Context) {
	models, err := h.userModelService.List(c.Request.Context())
	if err != nil {
		h.l.Error("list user models failed", zapx.Error(err))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, models)
}
