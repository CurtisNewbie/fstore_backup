package fbackup

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/curtisnewbie/miso/miso"
)

const (
	PageLimit                  = 100
	PropStorage                = "backup.storage"
	PropTrash                  = "backup.trash"
	PropLocalCopyEnabled       = "backup.local-copy.enabled"
	PropLocalCopyFstoreStorage = "backup.local-copy.fstore-storage"

	StatusNormal = "NORMAL"  // file.status - normal
	StatusLDel   = "LOG_DEL" // file.status - logically deletedy
	StatusPDel   = "PHY_DEL" // file.status - physically deletedy

	ThrottleUnit int64 = 7 * 1024 * 1024
)

func init() {
	miso.SetDefProp(PropStorage, "./storage")
	miso.SetDefProp(PropTrash, "./trash")
}

func storageDir() (string, error) {
	storageDir := miso.GetPropStr(PropStorage)
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to MkdirAll for stroage path, %v, %v", storageDir, err)
	}
	return storageDir, nil
}

func trashDir() (string, error) {
	trashDir := miso.GetPropStr(PropTrash)
	if err := os.MkdirAll(trashDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to MkdirAll for trash path, %v, %v", trashDir, err)
	}
	return trashDir, nil
}

func StartSync(rail miso.Rail) error {
	defer miso.TimeOp(rail, time.Now(), "Sync mini-fstore files")

	storageDir, err := storageDir()
	if err != nil {
		return err
	}

	trashDir, err := trashDir()
	if err != nil {
		return err
	}

	listReq := ListBackupFileReq{
		IdOffset: 0,
		Limit:    PageLimit,
	}

	var accBytes int64 = 0

	for {
		listed, err := ListFiles(rail, listReq)
		if err != nil {
			return fmt.Errorf("failed to list files, req: %+v, %v", listReq, err)
		}
		if len(listed.Files) < 1 {
			rail.Infof("Finished syncing all the files, lastId: %v", listReq.IdOffset)
			return nil
		}
		for i := range listed.Files {
			f := listed.Files[i]
			fetched, err := SyncFile(rail, f, storageDir, trashDir)
			if err != nil {
				return fmt.Errorf("failed to sync file, %v, %v", f, err)
			}

			if miso.IsShuttingDown() {
				rail.Info("server shutting down")
				return nil
			}

			if fetched {
				accBytes += f.Size
				if accBytes > ThrottleUnit {
					time.Sleep(time.Duration(int64(accBytes/ThrottleUnit)) * 100 * time.Millisecond)
					accBytes = 0
				}
			}

		}
		listReq.IdOffset = listed.Files[len(listed.Files)-1].Id
		rail.Infof("IdOffset moved to %v", listReq.IdOffset)
	}
}

func SyncFile(rail miso.Rail, bfi BackupFileInf, storageDir string, trashDir string) (bool, error) {
	rail.Infof("Sync file: %+v", bfi)

	spath := storageDir + "/" + bfi.FileId
	tpath := trashDir + "/" + bfi.FileId

	// file deleted
	if bfi.Status != StatusNormal {
		found, err := miso.FileExists(spath)
		if err != nil {
			return false, fmt.Errorf("failed to check if file exist, path: %v, %v", spath, err)
		}
		if !found {
			rail.Infof("File already deleted, already synced, %v", spath)
			return false, nil
		}
		if err := os.Rename(spath, tpath); err != nil {
			return false, fmt.Errorf("failed to move file from %v to %v, %v", spath, tpath, err)
		}
		rail.Infof("Moved file from %v to %v", spath, tpath)
		return false, nil
	}

	// the file should be downloaded, check if we have it already
	download := false
	fi, err := os.Stat(spath)

	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to stat file, path: %v, %v", spath, err)
		}
		rail.Infof("File %v is not found, downloading", spath)
		download = true // file is not found at all
	} else {
		// size doesn't match
		size := fi.Size()
		if size != bfi.Size {
			rail.Infof("File size doesn't match, downloading, path: %v, expected: %v, actual: %v", spath, bfi.Size, size)
			download = true
		} else {
			rail.Infof("File found and size matched, already synced, path: %v", spath)
		}
	}

	if !download {
		return false, nil
	}

	nf, err := os.Create(spath)
	if err != nil {
		return false, fmt.Errorf("failed to create file to download, path: %v, %v", spath, err)
	}

	doDownload := func() error {
		if err := DownloadFile(rail, bfi.FileId, nf); err != nil {
			return fmt.Errorf("failed to download file to %v, %v", spath, err)
		}
		return nil

	}

	if miso.GetPropBool(PropLocalCopyEnabled) {

		fstoreLocalStore := miso.GetPropStr(PropLocalCopyFstoreStorage)
		fstorePath := fstoreLocalStore + bfi.FileId
		found, err := miso.FileExists(fstorePath)
		if err != nil || !found {
			rail.Infof("Failed to access mini-fstore local file, fallback to file download, %v, %v", fstorePath, err)
			return true, doDownload()
		}

		fstoreFile, err := os.Open(fstorePath)
		if err != nil {
			rail.Infof("Failed to access mini-fstore local file, fallback to file download, %v, %v", fstorePath, err)
			return true, doDownload()
		}

		rail.Infof("Copying mini-fstore file directly from %v to %v", fstorePath, spath)
		_, err = io.Copy(nf, fstoreFile)
		return true, err
	} else {
		return true, doDownload()
	}
}
