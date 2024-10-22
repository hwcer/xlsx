package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/logger"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"strings"
)

func VerifyName(s string) (k string, ok bool) {
	i := strings.Index(s, ":")
	if i == -1 {
		return s, true
	}
	k = s[i+1:]
	tag := strings.ToUpper(cosgo.Config.GetString(FlagsNameTag))
	ok = tag == "" || tag == strings.ToUpper(s[0:i])
	return
}

// Convert 全角转半角
func Convert(s string) string {
	str := []rune(s)
	length := len(str)
	for i := 0; i < length; i++ {
		if str[i] == 12288 {
			str[i] = 32
			continue
		} else if str[i] > 65280 && str[i] < 65375 {
			str[i] = str[i] - 65248
		}
	}
	return string(str)
}

func TrimProtoName(s string) string {
	s = Convert(s)
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "_")
	s = FirstUpper(s)
	i := strings.Index(s, "_")
	for i > 0 {
		s = s[0:i] + FirstUpper(s[i+1:])
		i = strings.Index(s, "_")
	}
	return s
}

func Ignore(f string) bool {
	_, name := filepath.Split(f)
	if strings.HasPrefix(Convert(name), "~") {
		return false
	}
	if !strings.HasSuffix(f, ".xlsx") {
		return false
	}
	for _, v := range ignoreFiles {
		if strings.HasPrefix(f, v) {
			return false
		}
	}
	return true
}

func Valid(sheet *xlsx.Sheet) bool {
	r, e := sheet.Row(0)
	if e != nil {
		logger.Fatal("获取sheet行错误 name:%v,err:%v", sheet.Name, e)
	}
	cell := r.GetCell(0)
	return cell != nil && cell.Value != ""
}

func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func GetFiles(dir string, filter func(string) bool) (r []string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Fatal(err)
	}
	for _, info := range files {
		if info.IsDir() {
			r = append(r, GetFiles(filepath.Join(dir, info.Name()), filter)...)
		} else {
			f := filepath.Join(dir, info.Name())
			if filter(f) {
				r = append(r, f)
			}
		}
	}
	return
}

func preparePath() {
	var err error
	// excel文件必须存在
	logger.Trace("====================开始检查EXCEL路径====================")
	root := cosgo.Dir()
	in := cosgo.Config.GetString(FlagsNameIn)
	if !filepath.IsAbs(in) {
		in = filepath.Join(root, in)
	}
	if excelStat, err := os.Stat(in); err != nil || !excelStat.IsDir() {
		logger.Fatal("excel路径必须存在且为目录: %v ", in)
	}
	cosgo.Config.Set(FlagsNameIn, in)
	logger.Trace("输入目录:%v", in)

	logger.Trace("====================开始检查输出路径====================")
	if out := cosgo.Config.GetString(FlagsNameOut); out != "" {
		if !filepath.IsAbs(out) {
			out = filepath.Join(root, out)
		}
		if excelStat, err := os.Stat(out); err != nil || !excelStat.IsDir() {
			logger.Fatal("静态数据目录错误: %v ", out)
		}
		files, _ := os.ReadDir(out)
		logger.Trace("删除输出路径中的文件")
		for _, filename := range files {
			if strings.HasSuffix(filename.Name(), ".proto") ||
				strings.HasSuffix(filename.Name(), ".txt") {
				err := os.Remove(filepath.Join(out, filename.Name()))
				if err != nil {
					logger.Fatal(err)
				}
			}
		}
		cosgo.Config.Set(FlagsNameOut, out)
		logger.Trace("输出目录:%v", out)
	}
	logger.Trace("====================开始检查GO输出路径====================")
	if goOutPath := cosgo.Config.GetString(FlagsNameGo); goOutPath != "" {
		if !filepath.IsAbs(goOutPath) {
			goOutPath = filepath.Join(root, goOutPath)
		}
		if ext := filepath.Ext(goOutPath); ext != "" {
			err = checkFileAndRemove(goOutPath)
		} else {
			err = checkDirAndRemove(goOutPath, "")
		}
		if err != nil {
			logger.Fatal(err)
		}
		cosgo.Config.Set(FlagsNameGo, goOutPath)
		logger.Trace("GO输出目录:%v", goOutPath)
	}

	logger.Trace("====================开始检查JSON输出路径====================")
	if jsonPath := cosgo.Config.GetString(FlagsNameJson); jsonPath != "" {
		if !filepath.IsAbs(jsonPath) {
			jsonPath = filepath.Join(root, jsonPath)
		}
		if ext := filepath.Ext(jsonPath); ext != "" {
			err = checkFileAndRemove(jsonPath)
		} else {
			err = checkDirAndRemove(jsonPath, ".json")
		}
		if err != nil {
			logger.Fatal(err)
		}
		cosgo.Config.Set(FlagsNameJson, jsonPath)
		logger.Trace("JSON输出目录:%v", jsonPath)
	}
	logger.Trace("====================开始检查Info输出路径====================")
	if infoPath := cosgo.Config.GetString(FlagsNameInfo); infoPath != "" {
		if !filepath.IsAbs(infoPath) {
			infoPath = filepath.Join(root, infoPath)
		}
		if ext := filepath.Ext(infoPath); ext == "" {
			infoPath = filepath.Join(infoPath, "info.json")
		}
		if err = checkFileAndRemove(infoPath); err != nil {
			logger.Fatal(err)
		}
		cosgo.Config.Set(FlagsNameInfo, infoPath)
		logger.Trace("INFO输出目录:%v", infoPath)
	}

	logger.Trace("====================开始检查忽略文件列表====================")
	if s := cosgo.Config.GetString(FlagsNameIgnore); s != "" {
		for _, v := range strings.Split(s, ",") {
			if v != "" {
				f := filepath.Join(in, v)
				ignoreFiles = append(ignoreFiles, f)
				logger.Trace("忽略路径:%v", f)
			}

		}
	}
	logger.Trace("====================开始检查多语言文件====================")
	if languagePath := cosgo.Config.GetString(FlagsNameLanguage); languagePath != "" {
		if !filepath.IsAbs(languagePath) {
			languagePath = filepath.Join(root, languagePath)
		}
		if excelStat, err := os.Stat(languagePath); err != nil {
			logger.Fatal("语言文件错误: %v ", err)
		} else if excelStat.IsDir() {
			logger.Fatal("语言文件不能是一个目录: %v ", languagePath)
		} else if ext := filepath.Ext(languagePath); ext != ".xlsx" && ext != ".xls" {
			logger.Fatal("语言文件必须是Excel(xlsx,xls) ")
		}
		cosgo.Config.Set(FlagsNameLanguage, languagePath)
	} else {
		logger.Trace("未设置语言文件,已经跳过")
	}
}

func checkDirAndRemove(path string, ext string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("目录错误:%v", path)
	}
	fs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if ext == "" {
		return nil
	}
	logger.Trace("清理目录:%v", path)
	for _, filename := range fs {
		if strings.HasSuffix(strings.ToLower(filename.Name()), ext) {
			if err = os.Remove(filepath.Join(path, filename.Name())); err != nil {
				return fmt.Errorf("清理文件错误:%v", path)
			}
		}
	}
	return nil
}
func checkFileAndRemove(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	if stat.IsDir() {
		return fmt.Errorf("文件错误:%v", path)
	}
	return os.Remove(path)
}
