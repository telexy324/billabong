package upload

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/telexy324/billabong/service/singleton"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"
)

type Local struct{}

//@author: [piexlmax](https://github.com/piexlmax)
//@author: [ccfish86](https://github.com/ccfish86)
//@author: [SliverHorn](https://github.com/SliverHorn)
//@object: *Local
//@function: UploadFile
//@description: 上传文件
//@param: file *multipart.FileHeader
//@return: string, string, error

func (*Local) UploadFile(file *multipart.FileHeader) (string, string, error) {
	// 读取文件后缀
	ext := path.Ext(file.Filename)
	// 读取文件名并加密
	name := strings.TrimSuffix(file.Filename, ext)
	name = MD5V([]byte(name))
	// 拼接新文件名
	filename := name + "_" + time.Now().Format("20060102150405") + ext
	// 尝试创建此路径
	mkdirErr := os.MkdirAll(singleton.Conf.Location, os.ModePerm)
	if mkdirErr != nil {
		return "", "", singleton.Localizer.ErrorT("function os.MkdirAll() Filed, err:" + mkdirErr.Error())
	}
	// 拼接路径和文件名
	p := singleton.Conf.LocalPath + "/" + filename

	f, openError := file.Open() // 读取文件
	if openError != nil {
		return "", "", singleton.Localizer.ErrorT("function file.Open() Filed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭

	out, createErr := os.Create(p)
	if createErr != nil {
		return "", "", singleton.Localizer.ErrorT("function os.Create() Filed, err:" + createErr.Error())
	}
	defer out.Close() // 创建文件 defer 关闭

	_, copyErr := io.Copy(out, f) // 传输（拷贝）文件
	if copyErr != nil {
		return "", "", singleton.Localizer.ErrorT("function io.Copy() Filed, err:" + copyErr.Error())
	}
	return p, filename, nil
}

//@author: [piexlmax](https://github.com/piexlmax)
//@author: [ccfish86](https://github.com/ccfish86)
//@author: [SliverHorn](https://github.com/SliverHorn)
//@object: *Local
//@function: DeleteFile
//@description: 删除文件
//@param: key string
//@return: error

func (*Local) DeleteFile(key string) error {
	p := singleton.Conf.LocalPath + "/" + key
	if strings.Contains(p, singleton.Conf.LocalPath) {
		if err := os.Remove(p); err != nil {
			return singleton.Localizer.ErrorT("本地文件删除失败, err:" + err.Error())
		}
	}
	return nil
}

func MD5V(str []byte, b ...byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(b))
}
