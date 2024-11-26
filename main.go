package main

import (
	"os"
	"time"

	"github.com/curtisnewbie/fstore_backup/fbackup"
	"github.com/curtisnewbie/miso/miso"
)

func main() {

	miso.AddShutdownHook(func() { miso.MarkServerShuttingDown() }) // this is a bug
	miso.AddShutdownHook(func() { time.Sleep(2 * time.Second) })

	miso.PostServerBootstrapped(func(rail miso.Rail) error {
		go func() {
			if err := fbackup.StartSync(rail); err != nil {
				rail.Errorf("failed to sync, %v", err)
				return
			}
			miso.Shutdown()
		}()
		return nil
	})
	miso.BootstrapServer(os.Args)
}
