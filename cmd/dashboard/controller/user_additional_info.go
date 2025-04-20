package controller

import (
	"gorm.io/gorm"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/service/singleton"
)

// Get user additional info
// @Summary Get user additional info
// @Security BearerAuth
// @Schemes
// @Description Get user additional info
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[model.UserAdditionalInfo]
// @Router /user/additional/{id} [get]
func getUserAdditionalInfoById(c *gin.Context) (*model.UserAdditionalInfo, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var userAdditionalInfo *model.UserAdditionalInfo
	if err = singleton.DB.Where("id = ?", id).Find(userAdditionalInfo).Error; err != nil {
		return nil, newGormError("%v", err)
	}
	return userAdditionalInfo, nil
}

// Update password for current user
// @Summary Update password for current user
// @Security BearerAuth
// @Schemes
// @Description Update password for current user
// @Tags auth required
// @Accept json
// @param request body model.UserAdditionalForm true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /tool/{id} [patch]
func updateUserAdditionalInfo(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	var uf model.UserAdditionalForm
	if err = c.ShouldBindJSON(&uf); err != nil {
		return 0, err
	}

	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	var oldUserAdditionalInfo model.UserAdditionalInfo
	upDateMap := make(map[string]interface{})
	upDateMap["summary"] = uf.Avatar
	upDateMap["description"] = uf.Description

	err = singleton.DB.Transaction(func(tx *gorm.DB) error {
		db := tx.Where("id = ?", id).Find(&oldUserAdditionalInfo)
		txErr := db.Updates(upDateMap).Error
		if txErr != nil {
			return singleton.Localizer.ErrorT("update tool failed: %v", txErr)
		}
		return nil
	})
	return 0, err
}

// Create user additional info
// @Summary Create user additional info
// @Security BearerAuth
// @Schemes
// @Description Create user additional info
// @Tags admin required
// @Accept json
// @param request body model.UserAdditionalForm true "Tool Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /tool [post]
func createUserAdditionalInfo(c *gin.Context) (uint64, error) {
	var uf model.UserAdditionalForm
	var u model.UserAdditionalInfo
	if err := c.ShouldBindJSON(&uf); err != nil {
		return 0, err
	}
	uid := getUid(c)

	u.UserID = uid
	u.Avatar = uf.Avatar
	u.Description = uf.Description

	if err := singleton.DB.Create(&u).Error; err != nil {
		return 0, err
	}

	return u.ID, nil
}
