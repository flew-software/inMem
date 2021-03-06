/*
This file contains all base functions used by commands
*/

package commands

import (
	"fmt"
	"github.com/inancgumus/screen"
	"inMem/memory"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// downloads a file from a website into memory file system
func HttpGetToMem(memfs *memory.FileSystem, url string, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	_, err = memfs.CreateFile(fileName)
	if err != nil {
		return err
	}

	_, err = memfs.WriteFS(fileName, body)
	if err != nil {
		return err
	}

	return nil
}

// hosts data
func HostData(memfs *memory.FileSystem, location string, port int, pattern string) error {
	f, err := memfs.ReadFS(location)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f.Reader())
	if err != nil {
		return err
	}

	KeepHosting := true
	p := len(CommandProcesses)
	t := time.Now()
	CommandProcesses = append(CommandProcesses, CommandProcess{
		ProcessName: "Hosting " + location + " in port " + strconv.Itoa(port),
		Command:     GetCommands()["host"],
		KillFunc: func() {
			KeepHosting = false
			CommandProcesses[p].Killed = true
		},
		Killed:  false,
		Created: t.Unix(),
		End:     0,
	})

	defer func() {
		CommandProcesses[p].setKilled()
	}()

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if KeepHosting {
			_, err = w.Write(b)
			if err != nil {
				log.Fatalln(err)
			}
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

	return nil
}

// stashes the session in session store
func StashSession(memfs *memory.FileSystem, id string) {
	sessionStore[id] = memory.FileSystemStoreEntry{
		FS:     *memfs,
		Stored: time.Now().Unix(),
	}
	memfs.ClearFS()
}

// collects a session from the session store
func CollectSession(memfs *memory.FileSystem, id string, stashCurrent bool, newId string) {
	s := sessionStore[id]
	fs := s.FS.MFileSystem
	if stashCurrent {
		sessionStore[newId] = memory.FileSystemStoreEntry{
			FS:     *memfs,
			Stored: time.Now().Unix(),
		}
	}
	memfs.ReplaceFS(fs)
}

// clears screen
func ClearScreen() {
	screen.Clear()
	screen.MoveTopLeft()
}

func Kill() int {
	var killedProcesses int = 0

	for i := 0; i < len(CommandProcesses); i++ {
		if CommandProcesses[i].Killed == false {
			fmt.Printf("Killing %s child of command %s\n", CommandProcesses[i].ProcessName, CommandProcesses[i].Command.Prefix)
			CommandProcesses[i].KillFunc()
			CommandProcesses[i].Killed = true
			CommandProcesses[i].End = time.Now().Unix()
			killedProcesses++
		}
	}
	return killedProcesses
}
