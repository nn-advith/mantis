// utilites package

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// add logger here ig

func initLogger() {
	log.SetPrefix("mantis: ")
}

func LogProcessInfo(proc *os.Process, logval string) {
	log.SetPrefix(strconv.Itoa(proc.Pid) + "> ")
	log.SetFlags(0)
	log.Println(logval)
	log.SetFlags(1)
	log.SetPrefix("mantis: ")
}

func usage() {
	fmt.Printf("\nUsage:\n\nmantis -f <files>/<directory> -a <args> -e <key=value>\nmantis -v for version\nmantis -h for help")
}

func runtimeCommandsLegend() {
	fmt.Printf("\nAllowed command chars:\nq\t-\tQuit mantis\nr\t-\tRestart processes\n\n")
}

func returnVersion() {
	fmt.Println("mantis v1.0.0")
}

func ParseArgs(gargs map[string][]string, args []string) error {
	var indexes = map[string]int{"-a": -1, "-e": -1, "-f": -1}
	tmpargs := args[1:]
	currentkey := ""
	for i := range tmpargs {
		if _, exists := indexes[tmpargs[i]]; tmpargs[i][0] == '-' && !exists {
			return fmt.Errorf("unknown key for normal usage")
		}
		// add error check if flag is empty
		if _, exists := indexes[tmpargs[i]]; exists {
			currentkey = tmpargs[i]
			continue
		}
		gargs[currentkey] = append(gargs[currentkey], tmpargs[i])

	}
	return nil

}

func CleanFileArgs() error {
	if len(globalargs["-f"]) == 1 {
		f_arg := globalargs["-f"][0]
		info, err := os.Stat(f_arg)
		if err != nil {
			return fmt.Errorf("error during stat %v", err)
		}
		if info.IsDir() {
			pattern := `^[\.]?[\/]?[A-Za-z0-9_]+\/$`
			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(f_arg)

			if re.MatchString(f_arg) {
				f_arg = match[0] + "*.go"
				files, err := filepath.Glob(f_arg)
				if err != nil || len(files) == 0 {
					return fmt.Errorf("no go files found; halting execution")
				}
				globalargs["-f"] = files
			}
		}
	}
	return nil
}

func CheckForGlobalConfig() error {
	gcfilepath := GetGlobalConfigPath()
	if _, err := os.Stat(gcfilepath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(gcfilepath), 0755)
		if err != nil {
			return fmt.Errorf("error creating global config directory")
		}

		defaultConfig := map[string]string{
			"extensions": ".go",
			"ignore":     "",
			"delay":      "0",
			"env":        "",
			"args":       "",
		}
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return fmt.Errorf("error while marshalling global config")
		}
		err = os.WriteFile(gcfilepath, data, 0644)
		if err != nil {
			return fmt.Errorf("error writing default global config")
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("error during stat; global config")
	}
	return nil
}

func CheckForLocalConfig(ftags []string) (bool, error) {
	if s, err := os.Stat(ftags[0]); err != nil {
		return false, fmt.Errorf("error during stat")
	} else {
		if s.IsDir() {
			wdirectory, err = filepath.Abs(ftags[0])
			if err != nil {
				return false, fmt.Errorf("error while checking for local config")
			}
		} else {
			wdirectory, err = filepath.Abs(filepath.Dir(ftags[0]))
			if err != nil {
				return false, fmt.Errorf("error while checking for local config")
			}
		}
	}
	localconfig := filepath.Join(wdirectory, "mantis.json")
	if present, err := os.Stat(localconfig); err == nil {
		return true && !present.IsDir(), nil
	}
	return false, nil

}

func GetFilesToMonitor() {
	extensions := mantis_config.Extensions
	ignore := mantis_config.Ignore

	extlist := strings.Split(extensions, ",")
	ignorelist := strings.Split(ignore, ",")

	filepath.WalkDir(wdirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal("error while scanning files to monitor", err)
			return nil
		}

		if d.IsDir() {
			if d.Name()[0] == '.' {
				return filepath.SkipDir
			} else {
				if slices.Contains(ignorelist, d.Name()+"/") {
					return filepath.SkipDir
				}
			}
		} else {
			t := strings.Split(d.Name(), ".")
			ext := "." + t[len(t)-1]
			t_path := strings.Replace(filepath.ToSlash(path), filepath.ToSlash(wdirectory)+"/", "", -1)

			if slices.Contains(extlist, ext) && !slices.Contains(ignorelist, t_path) {

				fileinfo, _ := os.Stat(path)
				monitor_list[path] = []int{int(fileinfo.Size()), int(fileinfo.ModTime().Unix())}
			} else {
				//skip
			}
		}
		return nil
	})

}

func DecodeMantisConfig() error {
	file, err := os.Open(config_file)
	if err != nil {
		return fmt.Errorf("error while opening %s", config_file)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&mantis_config); err != nil {
		return fmt.Errorf("error while decoding mantis.json")
	}

	//flags is not handled as of now. ignored

	return nil
}

func PreExec() error {

	if len(os.Args) == 1 {
		usage()
		os.Exit(0)
	}
	if len(os.Args) == 2 {
		if os.Args[1] == "-h" {
			usage()
			os.Exit(0)
		}
		if os.Args[1] == "-v" {
			returnVersion()
			os.Exit(0)
		}
	}

	globalargs = map[string][]string{
		"-f": make([]string, 0),
		"-a": make([]string, 0),
		"-e": make([]string, 0),
	}
	err := ParseArgs(globalargs, os.Args)
	if err != nil {
		log.Printf("parse error: %v", err)
		usage()
		os.Exit(1)
	}
	err = CleanFileArgs()
	if err != nil {
		return err
	}

	err = CheckForGlobalConfig()
	if err != nil {
		return err
	}

	localconfigpresent, err := CheckForLocalConfig(globalargs["-f"])
	if err != nil {
		return err
	}
	if localconfigpresent {
		log.Printf("found local mantis config")
		config_file = filepath.Join(wdirectory, "mantis.json")
	} else {
		log.Printf("using global mantis config")
		config_file = GetGlobalConfigPath()
	}

	err = DecodeMantisConfig()
	if err != nil {
		return err
	}

	GetFilesToMonitor()

	return nil
}
