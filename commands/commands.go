/*
This file contains high level commands that is used by the user
*/

package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/awnumar/memguard"
	"inMem/internal_processes"
	"inMem/memory"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"
)

func DownloadCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	fileName := c[1]
	url := c[2]
	defer func() {
		if err := recover(); err != nil {
			log.Println("Download failed due to not having enough args:", err)
		}
	}()

	fmt.Printf("Downloading %s\n", url)
	err := HttpGetToMem(fs, url, *dir+"/"+fileName)
	if err != nil {
		log.Println(err)
	}
}

func HostCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("Hosting failed due to not having enough args:", err)
		}
	}()

	location := c[1]
	port, err := strconv.Atoi(c[2])
	pattern := c[3]

	if err != nil {
		log.Println("Port is not an number")
	} else {
		go func() {
			err := HostData(fs, *dir+"/"+location, port, pattern)
			if err != nil {
				fmt.Printf("cant host %s: %v\n", location, err)
			}
		}()
	}

}

func KillCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	killedProcesses := Kill()
	fmt.Printf("Killed %d procces(es)\n", killedProcesses)
}

func ListProcessesCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	var processCount = 0
	t := time.Now()
	fmt.Printf("Current time: %d sec\n", t.Unix())
	for i := 0; i < len(CommandProcesses); i++ {
		var runningFor int64
		if CommandProcesses[i].Deleted {
			continue
		}
		if CommandProcesses[i].End > 0 {
			runningFor = CommandProcesses[i].End - CommandProcesses[i].Created
		} else {
			runningFor = t.Unix() - CommandProcesses[i].Created
		}
		fmt.Printf("proccess: %s | Created: %d | Runing for: %d sec | Alive: %t \n", CommandProcesses[i].ProcessName, CommandProcesses[i].Created, runningFor, !CommandProcesses[i].Killed)
		processCount++

	}
}

func CleanProcessList(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	CommandProcesses.Clear()
}

func ExitCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	memguard.SafeExit(0)
}

func ChangeDirCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("change dir failed due to not having enough args:", err)
		}
	}()

	newDir := c[1]
	if newDir == "." {
		newDir = *dir
	}

	if f, err := fs.MFileSystem.Stat(newDir); err == nil && f.IsDir() == true {
		*dir = path.Clean(newDir)
	} else if err != nil {
		println("could not change directory")
	}
}

func MakeDirCommand(c []string, _ *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("mkdir failed due to not having enough args:", err)
		}
	}()
	newDir := c[1]

	err := fs.MFileSystem.Mkdir(newDir, os.ModeDir)
	if err != nil {
		log.Printf("err %s", err)
	}
}

func ListCommand(_ []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := fs.MFileSystem.ReadDir(*dir)
	w := tabwriter.NewWriter(os.Stdout, 1, 3, 3, ' ', 0)

	if err != nil {
		fmt.Printf("error could not list dir: %s\n", err)
	} else {
		fmt.Printf("found %d dir(s)/file(s) in %s\n", len(f), *dir)
		fmt.Fprintf(w, "name\tisDir\tsize\text\n")
		for i := 0; i < len(f); i++ {
			_, err = fmt.Fprintf(w, "%v\t%t\t%d bytes\t%s\n", f[i].Name(), f[i].IsDir(), f[i].Size(), path.Ext(f[i].Name()))
			if err != nil {
				fmt.Printf("error could not list dir: %s\n", err)
			}
		}
		w.Flush()
	}

}

func NewSessionCommand(_ []string, _ *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	*fs = memory.FileSystem{}
	*fs = memory.CreateMemoryFileSystem()
	fmt.Println("New session created")
}

func FileHashCommand(c []string, _ *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("FileHash failed due to not having enough args:", err)
		}
	}()

	file := c[1]
	hash, err := fs.GetHash(file)
	if err != nil {
		fmt.Printf("Unable to get hash of %s: %v", file, err)
	} else {
		fmt.Printf("Hash of %s : %s", file, hex.EncodeToString(hash))
	}
}

func HelpCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	p := GetCommands()
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for _, val := range p {
		_, err := fmt.Fprintf(w, "%s\t%s\n", val.Prefix, val.Description)
		if err != nil {
			fmt.Print(err)
		}
	}
	w.Flush()
}

func MakeFileCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("make file failed due to not having enough args:", err)
		}
	}()

	fileName := c[1]
	_, err := fs.CreateFile(*dir + "/" + fileName)
	if err != nil {
		fmt.Printf("Unable to create %s: %v", fileName, err)
	}
}

func CollectSessionCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("session collection failed due to not having enough args:", err)
		}
	}()

	id := c[1]
	CollectSession(fs, id, false, "")
	fmt.Print("current session dropped\n")
	fmt.Printf("collected session with id: %s\n", id)
	*dir = "/"
}

func StashSessionCommand(c []string, dir *string, fs *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	id := c[1]
	defer func() {
		if err := recover(); err != nil {
			log.Println("session stashing failed due to not having enough args:", err)
		}
	}()

	Kill()

	// stashes session with id
	StashSession(fs, id)

	fmt.Print("current session dropped\n")
	fmt.Printf("session stashed with id: %s\n", id)

	// sets the dir back to /
	*dir = "/"
}

func ListSessionsCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	w := tabwriter.NewWriter(os.Stdout, 3, 1, 1, ' ', 0)
	fmt.Fprint(w, "id\tstored_time\n")

	for key, val := range sessionStore {

		t := time.Unix(val.Stored, 0)
		createdTime := fmt.Sprintf("%d:%d %d/%d/%d", t.Hour(), t.Minute(), t.Day(), t.Month(), t.Year())
		_, err := fmt.Fprintf(w, "%s\t%s", key, createdTime)
		if err != nil {
			fmt.Printf("unable to list sessions: %v", err)
		}
	}
	err := w.Flush()
	if err != nil {
		fmt.Printf("unable to list sessions: %v", err)
	}
}

func ProcessOutCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()

	internal_processes.InterSTD.PrintToStdOut()
}

func ClearCommand(_ []string, _ *string, _ *memory.FileSystem, wg *sync.WaitGroup) {
	defer wg.Done()
	ClearScreen()
}
