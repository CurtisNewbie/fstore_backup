package fbackup

import (
	"os"
	"testing"

	"github.com/curtisnewbie/miso/miso"
)

func TestSyncFile(t *testing.T) {
	rail := preTest()
	miso.SetProp(PropTrash, "../trash")
	miso.SetProp(PropStorage, "../storage")
	defer os.RemoveAll("../trash")
	defer os.RemoveAll("../storage")

	storageDir, err := storageDir()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	trashDir, err := trashDir()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	err = SyncFile(rail, BackupFileInf{
		Id:     92,
		FileId: "file_794563461529600120059",
		Status: StatusNormal,
		Size:   75903,
	}, storageDir, trashDir)

	if err != nil {
		t.Logf("err: %v", err)
		t.FailNow()
	}
}
