package main

import (
	"os"

	"github.com/curtisnewbie/fstore_backup/fbackup"
	"github.com/curtisnewbie/miso/miso"
)

func main() {
	miso.PostServerBootstrapped(func(rail miso.Rail) error {
		if err := fbackup.StartSync(rail); err != nil {
			return err
		}
		miso.Shutdown()
		return nil
	})
	miso.BootstrapServer(os.Args)
}
