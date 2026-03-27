package fileMonitor

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// 事件类型，传给上层处理
type FileEvent struct {
	Path string
	Op   fsnotify.Op
}

// 监听器结构体
type FileWatcher struct {
	watcher   *fsnotify.Watcher
	rootPaths []string
	recursive bool

	eventCallback func(FileEvent)
	errorCallback func(error)

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// 初始化监听器
func NewFileWatcher(paths []string, recursive bool) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw := &FileWatcher{
		watcher:   w,
		rootPaths: paths,
		recursive: recursive,
		stopChan:  make(chan struct{}),
	}
	return fw, nil
}

// 设置回调
func (fw *FileWatcher) OnEvent(cb func(FileEvent)) {
	fw.eventCallback = cb
}
func (fw *FileWatcher) OnError(cb func(error)) {
	fw.errorCallback = cb
}

// 启动监听
func (fw *FileWatcher) Start() error {
	for _, root := range fw.rootPaths {
		if fw.recursive {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return fw.watcher.Add(path)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			if err := fw.watcher.Add(root); err != nil {
				return err
			}
		}
	}

	fw.wg.Add(1)
	go fw.watchLoop()
	return nil
}

// 停止监听
func (fw *FileWatcher) Stop() error {
	close(fw.stopChan)
	fw.wg.Wait()
	return fw.watcher.Close()
}

// 内部循环
func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()
	for {
		select {
		case <-fw.stopChan:
			return
		case event := <-fw.watcher.Events:
			// 如果新建的是目录并开启递归，加入监听
			if event.Op&fsnotify.Create == fsnotify.Create {
				fi, err := os.Stat(event.Name)
				if err == nil && fi.IsDir() && fw.recursive {
					_ = fw.watcher.Add(event.Name)
				}
			}
			if fw.eventCallback != nil {
				fw.eventCallback(FileEvent{Path: event.Name, Op: event.Op})
			}
		case err := <-fw.watcher.Errors:
			if fw.errorCallback != nil {
				fw.errorCallback(err)
			}
		}
	}
}
