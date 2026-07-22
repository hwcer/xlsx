package xlsx

import (
	"fmt"
	"strings"

	"github.com/hwcer/logger"
	"github.com/xuri/excelize/v2"
)

// xlsx 为每个公式单元格保存两份数据:公式本身,以及上次算出的结果(缓存值)。
// 读表只取缓存值,从不自己算公式 —— 工作簿由 Excel 存过时这没有问题。
//
// 但工作簿若由不计算公式的工具写出(openpyxl 等脚本改表后直接保存,或关闭了
// 自动重算),缓存会是空的,该列会整列静默导成空值/0,构建依旧全绿,直到数据在
// 游戏里露馅才被发现。calcFormula 为此兜底:对空单元格回查是否挂着公式,有则
// 就地求值补回。
//
// 缓存完好的表不进入求值分支,零额外开销。
//
// 求值能力不等于 Excel:内置函数覆盖不全(如 INDEX 对单行数组常量按 row_num
// 解释,与 Excel 的按列取值不符),跨工作表引用读的仍是对方的缓存值(同样是空的)。
// 因此只要有一格算不出来,整张表的数据就不可信 —— 此时**中止导表**要求人工用
// Excel 重新保存,而不是导出一份半对半错的数据:后者与最初那个 bug 是同一种病。
type formulaResult struct {
	filled int      //补算成功
	failed int      //求值失败
	sample []string //失败样本,用于报错
}

// calcFormula 补算本表中无缓存值的公式单元格,就地写回 this.rows。
func (this *Sheet) calcFormula() {
	if len(this.rows) == 0 {
		return
	}
	maxCol := this.formulaMaxCol()
	if maxCol == 0 {
		return
	}

	r := &formulaResult{}
	for i, row := range this.rows {
		for j := 0; j < maxCol; j++ {
			if j < len(row) && row[j] != "" {
				continue
			}
			axis := getCellName(j+1, i+1)
			formula, err := this.excel.GetCellFormula(this.SheetName, axis)
			if err != nil || formula == "" {
				continue
			}
			value, ok := this.calcFormulaCell(r, axis, formula)
			if !ok || value == "" {
				continue
			}
			// GetRows 会裁掉行尾空单元格,补算的值可能落在裁剪区之外。
			for len(this.rows[i]) <= j {
				this.rows[i] = append(this.rows[i], "")
			}
			this.rows[i][j] = value
			r.filled++
		}
	}

	if r.filled == 0 && r.failed == 0 {
		return
	}
	// 即使全部补算成功,也说明这张表被不算公式的工具存过,不是正常状态,必须报出来。
	logger.Alert("****************表格[%v%v]有%v个公式单元格缺少缓存值,已重新求值补回", this.FileName, this.SheetName, r.filled)
	if r.failed == 0 {
		logger.Alert("****************请用 Excel 打开 %v 重新保存,以免下次导表再次触发", this.FileName)
		return
	}
	for _, s := range r.sample {
		logger.Alert("----%v", s)
	}
	logger.Fatal("表格[%v%v]有%v个公式单元格无法求值,数据不完整,已中止导表。\n"+
		"请用 Excel 打开 %v,确认公式结果正确后重新保存(Excel 会写回公式缓存值),再重新导表。",
		this.FileName, this.SheetName, r.failed, this.FileName)
}

// calcFormulaCell 求单个公式单元格的值。第二个返回值为 false 表示求值失败。
func (this *Sheet) calcFormulaCell(r *formulaResult, axis, formula string) (string, bool) {
	fail := func(reason string) (string, bool) {
		r.failed++
		if len(r.sample) < 5 {
			r.sample = append(r.sample, fmt.Sprintf("%v!%v  =%v  -> %v", this.SheetName, axis, formula, reason))
		}
		return "", false
	}
	value, err := this.excel.CalcCellValue(this.SheetName, axis)
	if err != nil {
		return fail(err.Error())
	}
	// Excel 错误值(#N/A / #REF! ...)在 Excel 里也显示为错误,按求值失败处理。
	if strings.HasPrefix(value, "#") {
		return fail(value)
	}
	return value, true
}

// formulaMaxCol 返回需要检查的列数。
//
// GetRows 会裁掉每行末尾的空单元格,只按 rows 的长度扫描会漏掉「整列都是无缓存
// 公式」且位于表格最右侧的列,故再用 sheet dimension 兜底,两者取大。
func (this *Sheet) formulaMaxCol() (maxCol int) {
	for _, row := range this.rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}
	dimension, err := this.excel.GetSheetDimension(this.SheetName)
	if err != nil || dimension == "" {
		return maxCol
	}
	i := strings.LastIndex(dimension, ":")
	if i < 0 {
		return maxCol
	}
	col, _, err := excelize.CellNameToCoordinates(dimension[i+1:])
	if err == nil && col > maxCol {
		maxCol = col
	}
	return maxCol
}
