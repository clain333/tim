package logic

import (
	config2 "cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	r "cc.tim/client/redis"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"

	"mime/multipart"
	"net/http"
	"time"
)

func CheckFileHashLogic(hash, filename string) (*pkg2.Response, error) {
	var size uint64
	var realFileName string
	err := db.MysqlDb.QueryRow("select file_size,real_file_name from files where file_hash=?", hash).Scan(&size, &realFileName)
	if err != nil && err != sql.ErrNoRows {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}
	if err == nil {
		fid := pkg2.SnowflakeNode.Generate().Int64()
		k := kafka.NewInstanceMysql()
		err := k.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg("INSERT files (id, real_file_name, original_filename, file_hash, file_size) VALUES (?,?,?,?,?)", uint64(fid), realFileName, filename, hash, size)})
		if err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
		}
		filekey := fmt.Sprintf(model.FileKEY, fid)
		r.RedisClient.Set(r.Ctx, filekey, "1", 1*time.Minute)
		return pkg2.NewResponse(http.StatusOK, "上传成功", map[string]interface{}{"fid": fid, "exists": true}), nil

	}
	return pkg2.NewResponse(http.StatusOK, "文件没有重复", map[string]interface{}{"exists": false}), nil

}

// UploadFile 处理文件上传
func UploadFileLogic(filename string, fileHeader *multipart.FileHeader) (*pkg2.Response, error) {
	if fileHeader.Size > config2.Config.Upload.Maxsize {
		return pkg2.NewResponse(http.StatusBadRequest, "上传失败，文件过大", nil), nil
	}
	realname := pkg2.SnowflakeNode.Generate().String()
	f := pkg2.NewFileType(realname)
	err := f.Save(fileHeader)
	if err != nil {
		f.Delete()
		return pkg2.NewResponse(http.StatusOK, "上传失败", nil), err
	}

	hash, err := pkg2.HashFile(realname)
	if err != nil {
		f.Delete()
		return pkg2.NewResponse(http.StatusInternalServerError, "上传失败", nil), err

	}
	var rfilename uint64
	err = db.MysqlDb.QueryRow("SELECT real_file_name from files where file_hash=?", hash).Scan(&rfilename)
	if err != nil && err != sql.ErrNoRows {
		f.Delete()
		return pkg2.NewResponse(http.StatusInternalServerError, "上传失败", nil), err
	}
	if err == nil {
		realname = pkg2.UintTurnStr(rfilename)
		f.Delete()
	}
	fid := pkg2.SnowflakeNode.Generate().Int64()
	k := kafka.NewInstanceMysql()
	err = k.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(`INSERT files (id, real_file_name, original_filename, file_hash, file_size) values (?,?,?,?,?)`, uint64(fid), realname, filename, hash, fileHeader.Size)})

	if err != nil {
		f.Delete()
		return pkg2.NewResponse(http.StatusInternalServerError, "上传失败", nil), err
	}
	filekey := fmt.Sprintf(model.FileKEY, fid)
	r.RedisClient.Set(r.Ctx, filekey, "1", 1*time.Minute)
	return pkg2.NewResponse(http.StatusOK, "上传成功", map[string]interface{}{"fid": fid, "exists": true}), nil
}

// GetFileByUUID 根据UUID获取文件元数据
func GetFileLogic(fileID string, c *gin.Context) error {
	var filename string
	var ruploadname string
	err := db.MysqlDb.QueryRow("SELECT real_file_name,original_filename FROM files WHERE id = ?", fileID).Scan(&filename, &ruploadname)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusOK, pkg2.NewResponse(http.StatusInternalServerError, err.Error(), nil))
		return err
	}
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, pkg2.NewResponse(http.StatusBadRequest, "文件不存在", nil))
	}
	f := pkg2.NewFileType(filename)

	c.FileAttachment(f.Load(), ruploadname)
	return nil
}
