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

func logProcessInfo(logval string) {
	log.SetPrefix(strconv.Itoa(cprocess.Pid) + ": ")
	log.Println(logval)
	log.SetPrefix("mantis: ")
}

func usage() {
	fmt.Printf("\nUsage:\n\nmantis -f <files>/<directory> -a <args> -e <key=value>\n")
}

func runtimeCommandsLegend() {
	fmt.Printf("\nAllowed command chars:\nq\t-\tQuit mantis\nr\t-\tRestart processes\n\n")
}

func parseArgs(gargs map[string][]string, args []string) error {
	var indexes = map[string]int{"-a": -1, "-e": -1, "-f": -1}
	tmpargs := args[1:]
	currentkey := ""
	for i := range tmpargs {
		if _, exists := indexes[tmpargs[i]]; tmpargs[i][0] == '-' && !exists {
			return fmt.Errorf("unknown key")
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

func cleanFileArgs() error {
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

func checkForGlobalConfig() error {
	gcfilepath := getGlobalConfigPath()
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

func checkForLocalConfig(ftags []string) (bool, error) {
	if s, err := os.Stat(ftags[0]); err != nil {
		return false, fmt.Errorf("error during stat")
	} else {
		if s.IsDir() {
			WDIR, err = filepath.Abs(ftags[0])
			if err != nil {
				return false, fmt.Errorf("error while checking for local config")
			}
		} else {
			WDIR, err = filepath.Abs(filepath.Dir(ftags[0]))
			if err != nil {
				return false, fmt.Errorf("error while checking for local config")
			}
		}
	}
	localconfig := filepath.Join(WDIR, "mantis.json")
	if present, err := os.Stat(localconfig); err == nil {
		return true && !present.IsDir(), nil
	}
	return false, nil

}

func getFilesToMonitor() {
	extensions := MANTIS_CONFIG.Extensions
	ignore := MANTIS_CONFIG.Ignore

	extlist := strings.Split(extensions, ",")
	ignorelist := strings.Split(ignore, ",")

	filepath.WalkDir(WDIR, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("error while scanning files to monitor", err)
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
			t_path := strings.Replace(filepath.ToSlash(path), filepath.ToSlash(WDIR)+"/", "", -1)

			if slices.Contains(extlist, ext) && !slices.Contains(ignorelist, t_path) {

				fileinfo, _ := os.Stat(path)
				MONITOR_LIST[path] = []int{int(fileinfo.Size()), int(fileinfo.ModTime().Unix())}
			} else {
				//skip
			}
		}
		return nil
	})

}

func decodeMantisConfig() error {
	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		return fmt.Errorf("error while opening %s", CONFIG_FILE)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&MANTIS_CONFIG); err != nil {
		return fmt.Errorf("error while decoding mantis.json")
	}

	//flags is not handled as of now. ignored

	return nil
}

func preExec() error {

	globalargs = map[string][]string{
		"-f": make([]string, 0),
		"-a": make([]string, 0),
		"-e": make([]string, 0),
	}
	err := parseArgs(globalargs, os.Args)
	if err != nil {
		fmt.Println("parse error", err)
		usage()
		os.Exit(1)
	}
	err = cleanFileArgs()
	if err != nil {
		return err
	}

	err = checkForGlobalConfig()
	if err != nil {
		return err
	}

	localconfigpresent, err := checkForLocalConfig(globalargs["-f"])
	if err != nil {
		return err
	}
	if localconfigpresent {
		CONFIG_FILE = filepath.Join(WDIR, "mantis.json")
	} else {
		CONFIG_FILE = getGlobalConfigPath()
	}

	err = decodeMantisConfig()
	if err != nil {
		return err
	}

	getFilesToMonitor()

	return nil
}
