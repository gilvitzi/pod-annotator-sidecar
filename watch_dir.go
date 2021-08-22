package main

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog"
)

func watchForChanges(wg *sync.WaitGroup, watcher *fsnotify.Watcher, changedFiles chan string) {
	defer wg.Done()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				klog.V(5).Info("watchForChanges finsihed")
				return
			}

			klog.V(5).Infof("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				klog.V(3).Infof("modified file:", event.Name)
				changedFiles <- event.Name
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				klog.V(3).Infof("new file:", event.Name)
				changedFiles <- event.Name
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				klog.V(3).Infof("removed file:", event.Name)
				changedFiles <- event.Name
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				klog.V(3).Infof("renamed file:", event.Name)
				changedFiles <- event.Name
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			klog.Info("error:", err)
		}
	}

}

func startWatchingDirForChanges(wg *sync.WaitGroup, directoryPath string) (*fsnotify.Watcher, chan string) {
	klog.V(5).Info("startWatchingDirForChanges")
	changedFiles := make(chan string, 1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatal(err)
	}

	wg.Add(1)
	go watchForChanges(wg, watcher, changedFiles)

	err = watcher.Add(directoryPath)
	if err != nil {
		klog.Fatal(err)
	}

	return watcher, changedFiles
}
