// utilites package

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// add logger here ig

func parseArgs(gargs map[string][]string, args []string) error {
	// find the tags locations
	var indexes = map[string]int{"-a": -1, "-e": -1, "-f": -1}
	tmpargs := args[1:]
	currentkey := ""
	for i := range tmpargs {

		if _, exists := indexes[tmpargs[i]]; tmpargs[i][0] == '-' && !exists {
			return fmt.Errorf("unknown key; refer usage")
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

func checkForGlobalConfig() error {
	gcfilepath := getGlobalConfigPath()
	if _, err := os.Stat(gcfilepath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(gcfilepath), 0755)

		defaultConfig := map[string]string{
			"monitor":    ".",
			"extensions": ".go",
			"main":       "go run main.go",
			"ignore":     "",
			"delay":      "0",
			"env":        "",
			"flags":      "",
		}
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return fmt.Errorf("error writing default global config : %v", err)

		}
		err = os.WriteFile(gcfilepath, data, 0644)
		if err != nil {
			return fmt.Errorf("error writing default global config")
		}
		return nil
	} else if err != nil {
	}
	return nil
}

func checkForLocalConfig(ftags []string) (bool, error) {
	// assume ftags has only one directory; it can be $/dir/ or $/dir/a.go
	// so check if it is dir, if dir move to checking for config
	// else find parent dir and then check for config
	fmt.Println(ftags[0])
	if s, err := os.Stat(ftags[0]); err != nil {
		return false, fmt.Errorf("error during stat %v", err)
	} else {
		if s.IsDir() {
			WDIR, _ = filepath.Abs(ftags[0])
		} else {
			WDIR, _ = filepath.Abs(filepath.Dir(ftags[0]))
		}
	}
	localconfig := filepath.Join(WDIR, "mantis.json")
	if present, err := os.Stat(localconfig); err == nil {
		return true && !present.IsDir(), nil
	}
	fmt.Println(WDIR)
	return false, nil

}

// array of filepaths to monitor. additionally
// map file size and mod time for each filepath

func getFilesToMonitor() {
	extensions := MANTIS_CONFIG.Extensions
	ignore := MANTIS_CONFIG.Ignore

	extlist := strings.Split(extensions, ",")
	ignorelist := strings.Split(ignore, ",")

	filepath.WalkDir(WDIR, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
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
			t_path := strings.Replace(filepath.ToSlash(path), WDIR, "", -1)

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

	fmt.Println(globalargs)

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
