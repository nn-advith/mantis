package filewatcher

import (
	"fmt"
	"os"
)

//for afile..get last mod time, last access time

//check size first. if size change, then definitley file changed.

// change this so that it returns modtime, err

func CheckFileExists(filepath string) error {
	fmt.Println(filepath)
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file doesn't exist: %s", err)
	}
	return nil
}

func GetModTime(filepath string) (int, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return 0, fmt.Errorf("error stat %v", err)
	}
	// fmt.Printf("%+v", fileInfo)
	return int(fileInfo.ModTime().Unix()), nil

}

func GetFileSize(filepath string) (int, error) {
	fileinfo, err := os.Stat(filepath)
	if err != nil {

		return 0, fmt.Errorf("error stat %v", err)

	}
	return int(fileinfo.Size()), nil
}

func GetFileHash(filepath string) string {
	return "HASH"
}

// func main() {
// 	fmt.Println("File watcher implementation")
// 	// fmt.Println(runtime.GOOS)
// 	fmt.Printf("%v", getModTime("./sample.txt"))
// }
