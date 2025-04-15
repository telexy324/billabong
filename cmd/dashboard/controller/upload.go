package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/pkg/upload"
	"github.com/telexy324/billabong/service/singleton"
	"strings"
)

// Upload file
// @Summary Upload file
// @Security BearerAuth
// @Schemes
// @Description Upload file
// @Tags admin required
// @Accept json
// @param file formData file true "Upload File"
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.Upload]
// @Router /file [post]
func uploadFile(c *gin.Context) (*model.Upload, error) {
	//noSave := c.DefaultQuery("noSave", "0")
	_, header, err := c.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	oss := upload.NewOss()
	filePath, key, uploadErr := oss.UploadFile(header)
	if uploadErr != nil {
		return nil, uploadErr
	}
	//if noSave == "0" {
	uid := getUid(c)
	s := strings.Split(header.Filename, ".")
	var f model.Upload
	f.UserID = uid
	f.Url = filePath
	f.Name = header.Filename
	f.Tag = s[len(s)-1]
	f.Key = key
	f.Size = header.Size
	if err = singleton.DB.Create(&f).Error; err != nil {
		return nil, err
	}
	//}
	return &f, nil
}

// Batch delete files
// @Summary Batch delete files
// @Security BearerAuth
// @Schemes
// @Description Batch delete files
// @Tags admin required
// @Accept json
// @param request body []uint true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/file [post]
func deleteFile(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}
	_, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}

	var files []model.Upload
	if err := singleton.DB.Where("id in ?", ids).Find(&files).Error; err != nil {
		return nil, err
	}
	oss := upload.NewOss()
	for _, file := range files {
		if err := oss.DeleteFile(file.Key); err != nil {
			return nil, err
		}
	}
	return nil, singleton.DB.Delete(&[]model.Tool{}, "id in ?", ids).Error
}
