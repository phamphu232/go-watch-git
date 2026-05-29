package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/kardianos/service"
	"github.com/phamphu232/go-watch-git/config"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	log.Printf("watch-git: start service")
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	log.Printf("watch-git: stop service")
	return nil
}

func (p *program) Restart(s service.Service) error {
	log.Printf("watch-git: restart service")
	return nil
}

func (p *program) Install(s service.Service) error {
	log.Printf("watch-git: install service")
	return nil
}

func (p *program) Uninstall(s service.Service) error {
	log.Printf("watch-git: uninstall service")
	return nil
}

func (p *program) run() {
	exePath, _ := os.Executable()
	lockPath := filepath.Join(filepath.Dir(exePath), ".pid.lock")
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	os.Chmod(lockPath, 0666)
	if err != nil || !locked {
		log.Println("watch-git is already running...")
		os.Exit(1)
	}

	releaseLock := func() {
		fileLock.Unlock()
		os.Remove(lockPath)
	}

	defer releaseLock()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		releaseLock()
		os.Exit(0)
	}()

	config.WatchConfig(3 * time.Second)

	runCheck()
}
func runCheck() {

	checkSourceCode()

	for {
		interval := config.GetConfig().Interval

		if interval == 0 && service.Interactive() {
			log.Println("watch-git stopped because interval = 0")
			break
		}

		if interval == 0 {
			time.Sleep(60 * time.Second)
		} else {
			time.Sleep(time.Duration(interval) * time.Second)
		}

		checkSourceCode()
	}

}

func main() {
	config.Load()
	initLogger()
	startCleanupWorker()

	exePath, _ := os.Executable()

	svcConfig := &service.Config{
		Name:        "watch-git",
		DisplayName: "watch-git",
		Description: "watch-git",

		WorkingDirectory: filepath.Dir(exePath),
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if service.Interactive() && len(os.Args) > 1 {
		if isAdmin() {
			service.Control(s, os.Args[1])
		} else {
			log.Printf("Please '%s' program as root/admin", os.Args[1])
		}

		return
	}

	err = s.Run()
	if err != nil {
		log.Println(err)
	}
}
