package fbackup

import (
	"os"
	"testing"

	"github.com/curtisnewbie/miso/miso"
)

func preTest() miso.Rail {
	rail := miso.EmptyRail()
	miso.DefaultReadConfig([]string{"configFile=../app-conf-dev.yml"}, rail)
	return rail
}

func TestListFiles(t *testing.T) {
	rail := preTest()
	r, err := ListFiles(rail, ListBackupFileReq{
		IdOffset: 84,
		Limit:    10,
	})
	if err != nil {
		t.Logf("err: %v", err)
		t.FailNow()
	}
	t.Logf("%+v", r)
}

func TestDownloadFile(t *testing.T) {
	rail := preTest()

	fname := "test_tmp"
	fileId := "file_794563461529600120059"

	f, err := os.Create(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(fname)

	err = DownloadFile(rail, fileId, f)
	if err != nil {
		t.Fatal(err)
	}
}
