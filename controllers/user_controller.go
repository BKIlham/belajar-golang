package controllers

import (
	"cobago/models"
	"cobago/services"
	"cobago/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	userService    services.UserService
	storageService services.StorageService
}

func NewUserController(us services.UserService, ss services.StorageService) *UserController {
	return &UserController{userService: us, storageService: ss}
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name")
	email := ctx.FormValue("email")
	password := ctx.FormValue("password")

	if name == "" || email == "" || password == "" {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Nama, email, password wajib diisi", nil)
	}

	fileHeader, err := ctx.FormFile("avatar")
	var avatarURL string
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal membaca file", err.Error())
		}
		defer file.Close()

		avatarURL, err = c.storageService.UploadAvatar(ctx.Context(), fileHeader.Filename, file, fileHeader.Size)
		if err != nil {
			return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal mengupload avatar ke MinIO", err.Error())
		}
	}

	user, err := c.userService.RegisterUser(name, email, password, avatarURL)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal menyimpan user ke database", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusCreated, "User berhasil diregistrasi", user)
}

func (c *UserController) GetProfile(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		return utils.SendError(ctx, fiber.StatusNotFound, "User tidak ditemukan", err.Error())
	}

	if user.AvatarURL != "" {
		secureURL, err := c.storageService.GetPresignedURL(ctx.Context(), user.AvatarURL)
		if err == nil {
			user.AvatarURL = secureURL
		}
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Data profil berhasil diambil", user)
}

func (c *UserController) Update(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	name := ctx.FormValue("name")
	email := ctx.FormValue("email")

	fileHeader, err := ctx.FormFile("avatar")
	var avatarURL string
	var inputData map[string]interface{}
	if err := ctx.BodyParser(&inputData); err != nil {
		oldUser, errFetch := c.userService.GetUserByID(uint(id))
		if errFetch == nil && oldUser.AvatarURL != "" {
			_ = c.storageService.DeleteFile(ctx.Context(), oldUser.AvatarURL)
		}

		file, _ := fileHeader.Open()
		defer file.Close()
		avatarURL, err = c.storageService.UploadAvatar(ctx.Context(), fileHeader.Filename, file, fileHeader.Size)
		if err != nil {
			return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal mengupload avatar baru", err.Error())
		}
	}

	user, err := c.userService.UpdateUser(uint(id), name, email, avatarURL)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal mengupdate user", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "User berhasil diperbarui", user)
}

func (c *UserController) SoftDelete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	if err := c.userService.SoftDeleteUser(uint(id)); err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal melakukan soft delete", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "User berhasil di-soft delete", nil)
}

func (c *UserController) HardDelete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	user, errFetch := c.userService.GetUserByID(uint(id))
	if errFetch == nil && user.AvatarURL != "" {
		_ = c.storageService.DeleteFile(ctx.Context(), user.AvatarURL)
	}

	if err := c.userService.HardDeleteUser(uint(id)); err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal melakukan hard delete", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "User dan file avatar berhasil dihapus permanen dari sistem", nil)
}

func (c *UserController) Restore(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "ID tidak valid", nil)
	}

	if err := c.userService.RestoreUser(uint(id)); err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal mengembalikan data user", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "User berhasil di-restore kembali", nil)
}

func (c *UserController) GetAll(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	// Teruskan ctx.Context() ke dalam service
	users, err := c.userService.GetAllUsers(ctx.Context(), page, limit)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal mengambil daftar user", err.Error())
	}

	if users == nil {
		users = []models.User{}
	}

	for i := range users {
		if users[i].AvatarURL != "" {
			secureURL, errLink := c.storageService.GetPresignedURL(ctx.Context(), users[i].AvatarURL)
			if errLink == nil {
				users[i].AvatarURL = secureURL
			}
		}
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Daftar user berhasil diambil", users)
}
