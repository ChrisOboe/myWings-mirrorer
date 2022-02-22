package main

import (
	"bufio"
	"fmt"
	"github.com/ChrisOboe/dirs"
	"github.com/ChrisOboe/wings-mirror/wings"
	"github.com/jessevdk/go-flags"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func cleanString(s string) string {
	var badChars = []string{"\\", "/", ":"}

	out := s
	for _, badChar := range badChars {
		out = strings.Replace(out, badChar, "_", -1)
	}

	return out
}

type Settings struct {
	Username    string `short:"u" long:"user" description:"The username of wings" required:"true"`
	Password    string `short:"p" long:"password" description:"The users password" required:"true"`
	Filepath    string `short:"o" long:"filename" description:"The filename used for downloding files"`
	Dbpath      string `short:"c" long:"cache" description:"The cachefile to be used"`
	PostCommand string `short:"m" long:"postcommand" description:"A command to run after the file was downloaded."`
}

func loadCache(path string) (map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return map[string]bool{}, fmt.Errorf("Couldn't open %s: %w", path, err)
	}
	defer file.Close()
	out := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out[scanner.Text()] = true
	}
	return out, nil
}

func writeCache(path string, cache map[string]bool) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Couldn't create %s: %w", filepath.Dir(path), err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Couldn't open %s: %w", path, err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for key := range cache {
		_, err := w.WriteString(key + "\n")
		if err != nil {
			return fmt.Errorf("Couldn't write to %s: %w", path, err)
		}
	}
	w.Flush()
	return nil
}

func main() {
	defaults := dirs.Get("wings-mirror")

	defaultFilepath := defaults.Home + "documents/%programName%/%courseName%/%fileTitle%"
	defaultCachepath := defaults.Cache + "cache"

	var s Settings
	parser := flags.NewParser(&s, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
	}

	if s.Filepath == "" {
		s.Filepath = defaultFilepath
	}
	if s.Dbpath == "" {
		s.Dbpath = defaultCachepath
	}

	cache, err := loadCache(s.Dbpath)
	if err != nil {
		fmt.Println(err)
	}

	w := wings.NewWings()
	err = w.Login(s.Username, s.Password)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Getting Stuidengaenge")
	programs, err := w.App.Programs()
	if err != nil {
		fmt.Println(err)
	}

	for _, p := range programs.Programs {
		fmt.Printf("Getting Courses for %s\n", p.Name)
		semesters, err := w.App.Semesters(strconv.Itoa(p.ID))
		if err != nil {
			fmt.Println(err)
		}

		for _, c := range semesters.Courses {
			fmt.Printf("Getting Data for %s\n", c.Name)
			module, err := w.MyWings.Modules(strconv.Itoa(p.ID), strconv.Itoa(c.PermanentID))
			if err != nil {
				fmt.Println(err)
			}

			for _, f := range module.Files {
				// check if we already downloaded this file
				cacheKey := f.Link + "/" + strconv.FormatInt(f.UpdatedAt.Unix(), 10)
				_, exists := cache[cacheKey]
				if exists {
					fmt.Printf("File %s was already downloaded. Skipping.\n", f.Title)
					continue
				}

				// resolve filename
				path := s.Filepath
				path = strings.Replace(path, "%programName%", cleanString(p.Name), -1)
				path = strings.Replace(path, "%courseName%", cleanString(c.Name), -1)
				path = strings.Replace(path, "%fileName%", cleanString(f.FileNameWithExtension), -1)
				path = strings.Replace(path, "%fileTitle%", cleanString(f.Title), -1)
				if !strings.Contains(s.Filepath, "%fileName%") {
					path = path + "." + f.Type
				}

				fmt.Printf("Getting file %s. Saving to %s\n", f.Title, path)
				err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = w.MyWings.Download(f.Link, path)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if s.PostCommand != "" {
					args := strings.Split(s.PostCommand, " ")
					args = append(args, path)
					cmd := exec.Command(args[0], args[1:]...)
					err := cmd.Run()
					if err != nil {
						fmt.Println(err)
						continue
					}
				}

				// update cache
				cache[cacheKey] = true
				writeCache(s.Dbpath, cache)
			}
		}
	}
}
