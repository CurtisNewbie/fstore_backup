package main

import (
	"os"

	"github.com/curtisnewbie/fstore_backup/fbackup"
	"github.com/curtisnewbie/miso/miso"
)

func main() {
	miso.PostServerBootstrapped(func(rail miso.Rail) error {
		go func() {
			if err := fbackup.StartSync(rail); err != nil {
				rail.Errorf("failed to sync, %v", err)
				miso.Shutdown()
				return
			}
		}()
		return nil
	})
	miso.BootstrapServer(os.Args)
}
