package flatfs

import (
	"metrics"
	"os"
	"runtime"
	"time"
)

// don't block more than 16 threads on sync opearation
// 16 should be able to sataurate most RAIDs
// in case of two used disks per write (RAID 1, 5) and queue depth of 2,
// 16 concurrent Sync calls should be able to saturate 16 HDDs RAID
//TODO: benchmark it out, maybe provide tweak parmeter
const SyncThreadsMax = 16

var syncSemaphore chan struct{} = make(chan struct{}, SyncThreadsMax)

func syncDir(dir string) error {
	s := time.Now()
	defer func() {
		if metrics.CMD_EnableMetrics {
			metrics.SyncTime.UpdateSince(s)
		}
	}()
	if runtime.GOOS == "windows" {
		// dir sync on windows doesn't work: https://git.io/vPnCI
		return nil
	}

	dirF, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer dirF.Close()

	syncSemaphore <- struct{}{}
	defer func() { <-syncSemaphore }()

	if err := dirF.Sync(); err != nil {
		return err
	}
	return nil
}

func syncFile(file *os.File) error {
	s := time.Now()
	defer func() {
		if metrics.CMD_EnableMetrics {
			metrics.SyncTime.UpdateSince(s)
		}
	}()
	syncSemaphore <- struct{}{}
	defer func() { <-syncSemaphore }()
	return file.Sync()
}
