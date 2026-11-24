package xlsx

import (
	"strings"
	"testing"
)

// EXCEL 不变时生成的 PROTO 必须逐字节一致:表顺序、总表编号、子对象 message 顺序都不能随机。

// 表按名字排序(不区分大小写),总表编号按排序后的顺序分配
func TestSortSheets(t *testing.T) {
	sheets := []*Sheet{
		{ProtoName: "Item"},
		{ProtoName: "hero"},
		{ProtoName: "Buff"},
		{ProtoName: "Skill"},
	}
	sortSheets(sheets)
	var names []string
	for _, s := range sheets {
		names = append(names, s.ProtoName)
	}
	if got, want := strings.Join(names, ","), "Buff,hero,Item,Skill"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, s := range sheets {
		if s.ProtoIndex != i+1 {
			t.Fatalf("%v ProtoIndex=%v, want %v", s.ProtoName, s.ProtoIndex, i+1)
		}
	}
}

// 子对象 message 按名字排序输出;globalObjects 是 map,重复多次验证不是碰巧有序
func TestBuildGlobalObjectsSorted(t *testing.T) {
	old, oldNamed := globalObjects, Config.NamedDummyInHeader
	t.Cleanup(func() { globalObjects, Config.NamedDummyInHeader = old, oldNamed })
	Config.NamedDummyInHeader = false

	var first string
	for i := 0; i < 50; i++ {
		globalObjects = GlobalDummy{}
		for _, name := range []string{"Delta", "Alpha", "Charlie", "Bravo", "Echo"} {
			d := NewDummy(name)
			if err := d.Add("id", ProtoBuffTypeInt32, 0); err != nil {
				t.Fatal(err)
			}
			d.Compile()
			globalObjects[d.Name] = d
		}
		b := &strings.Builder{}
		buildGlobalObjects(b, nil)
		out := b.String()
		pos := -1
		for _, name := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"} {
			p := strings.Index(out, "message "+name+"{")
			if p <= pos {
				t.Fatalf("子对象未按名字排序输出:\n%v", out)
			}
			pos = p
		}
		if i > 0 && out != first {
			t.Fatalf("第%d次输出与第1次不同:\n%v\n----\n%v", i+1, first, out)
		}
		first = out
	}
}
