package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"os"
	"os/exec"
	"path/filepath"
)

func ProtoGo() {
	logger.Trace("======================开始生成GO Message======================")
	out := fmt.Sprintf("--go_out=%v", cosgo.Config.GetString(FlagsNameGo))
	path := fmt.Sprintf("--proto_path=%v", cosgo.Config.GetString(FlagsNameOut))
	file := filepath.Join(cosgo.Config.GetString(FlagsNameOut), "*.proto")

	if err := os.Chdir(cosgo.Dir()); err != nil {
		logger.Fatal(err)
	}
	proto := cosgo.Config.GetString("protoc")
	if proto == "" {
		proto = joinPath("protoc")
	}
	plugin := cosgo.Config.GetString("protoc_plugin")
	if plugin == "" {
		plugin = fmt.Sprintf("--plugin=protoc-gen-go=%v", joinPath("protoc-gen-go.exe"))
	}

	//protoc --go_out=. --plugin=protoc-gen-go=path/to/protoc-gen-go your_proto_file.proto

	cmd := exec.Command(proto, out, plugin, path, file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Trace("Proto GO Path:%v", cosgo.Config.GetString(FlagsNameGo))
}

func joinPath(p string) string {
	appBinFile, _ := exec.LookPath(os.Args[0])
	var path string
	if filepath.IsAbs(appBinFile) {
		path = filepath.Dir(appBinFile)
	} else {
		workDir, _ := os.Getwd()
		path = filepath.Join(workDir, filepath.Dir(appBinFile))
	}
	return filepath.Join(path, p)
}
