package fbackup

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/curtisnewbie/miso/miso"
)

const (
	PropSecret     = "mini-fstore.secret"
	PropBaseUrl    = "mini-fstore.base-url"
	QryParamFileId = "fileId"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

type BackupFileInf struct {
	Id     int64
	FileId string
	Name   string
	Status string
	Size   int64
	Md5    string
}

type ListBackupFileReq struct {
	Limit    int
	IdOffset int64
}

func (l *ListBackupFileReq) Move(lastId int64) {
	l.IdOffset = lastId
}

type ListBackupFileResp struct {
	Files []BackupFileInf
}

func ListFiles(rail miso.Rail, req ListBackupFileReq) (ListBackupFileResp, error) {
	var resp miso.GnResp[ListBackupFileResp]
	err := miso.NewTClient(rail, miso.GetPropStr(PropBaseUrl)+"/fstore/backup/file/list").
		AddHeader("Authorization", miso.GetPropStr(PropSecret)).
		PostJson(req).
		Json(&resp)
	if err != nil {
		return ListBackupFileResp{}, err
	}
	return resp.Res()
}

func DownloadFile(rail miso.Rail, fileId string, writer io.Writer) error {
	r := miso.NewTClient(rail, miso.GetPropStr(PropBaseUrl)+"/fstore/backup/file/raw").
		AddHeader("Authorization", miso.GetPropStr(PropSecret)).
		AddQueryParams(QryParamFileId, fileId).
		Get()

	if r.Err != nil {
		return fmt.Errorf("unable to download file, fileId: %v, %v", fileId, r.Err)
	}
	if r.StatusCode == http.StatusNotFound {
		return fmt.Errorf("unable to download file, file is not found (404), fileId: %v, %w", fileId, ErrFileNotFound)
	}
	_, err := io.Copy(writer, r.Resp.Body)
	return err
}
