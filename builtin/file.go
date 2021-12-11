package builtin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	

)

func GetBuiltIn() (map[string]Builtin, error) {
	builtins := make(map[string]Builtin)

	
		for _, rt := range Basicbuiltins {
			builtins[rt.Name] = rt
		}

	
	return builtins, fmt.Errorf("something went worng")
}



func CreateDockerFile(internalPort int64, builtInFile Builtin, appName string)(string,error){

	if internalPort != 8080 {
		builtInFile.Template = strings.Replace(builtInFile.Template, "8080", fmt.Sprintf("%v", internalPort), -1)
	}

	getPath,err := FindPath("extracted_file/"+appName)

	if err != nil {
		log.Println(err)
		return "",err
	}

	filePath, _ := filepath.Abs("extracted_file/"+appName+"/"+getPath+"/Dockerfile")
	createfile, _ := os.Create(filePath)
	createfile.Close()
	err = ioutil.WriteFile(filePath, []byte(builtInFile.Template), 0644)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return "", err
}

func FindPath(filePath string)(string,error){
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
    var pathName string
	for _, v := range files {
		pathName = v.Name()
	}

	return pathName,err
}