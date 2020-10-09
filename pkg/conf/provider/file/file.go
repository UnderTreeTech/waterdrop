/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package file

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	fsnotify "gopkg.in/fsnotify.v1"
)

type fileProvider struct {
	path  string
	watch bool
}

func NewFileProvider(path string, watch bool) *fileProvider {
	return &fileProvider{path: filepath.Clean(path), watch: watch}
}

func (f *fileProvider) IsEnableWatch() bool {
	return f.watch
}

func (f *fileProvider) SetEnableWatch(enable bool) {
	f.watch = enable
}

func (f *fileProvider) ReadBytes() ([]byte, error) {
	return ioutil.ReadFile(f.path)
}

func (f *fileProvider) Watch(cb func()) error {
	// Resolve symlinks and save the original path so that changes to symlinks
	// can be detected.
	realPath, err := filepath.EvalSymlinks(f.path)
	if err != nil {
		return err
	}
	realPath = filepath.Clean(realPath)

	// Although only a single file is being watched, fsnotify has to watch
	// the whole parent directory to pick up all events such as symlink changes.
	fDir, _ := filepath.Split(f.path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	var (
		lastEvent     string
		lastEventTime time.Time
	)

	go func() {
	loop:
		for {
			select {
			case event := <-w.Events:
				log.Printf("event is %s", event.String())
				// Use a simple timer to buffer events as certain events fire
				// multiple times on some platforms.
				if event.String() == lastEvent && time.Since(lastEventTime) < time.Millisecond*5 {
					continue
				}
				lastEvent = event.String()
				lastEventTime = time.Now()

				evFile := filepath.Clean(event.Name)

				// Since the event is triggered on a directory, is this
				// one on the file being watched?
				if evFile != realPath && evFile != f.path {
					continue
				}

				// The file was removed.
				if event.Op&fsnotify.Remove != 0 {
					break loop
				}

				// Resolve symlink to get the real path, in case the symlink's
				// target has changed.
				curPath, err := filepath.EvalSymlinks(f.path)
				if err != nil {
					break loop
				}
				realPath = filepath.Clean(curPath)

				// Finally, we only care about create and write.
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}

				// Trigger event.
				cb()

			// There's an error.
			case err := <-w.Errors:
				log.Printf("watch file error, err msg %s", err.Error())
			}
		}

		w.Close()
	}()

	// Watch the directory for changes.
	return w.Add(fDir)
}
