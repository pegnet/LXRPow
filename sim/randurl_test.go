package sim

import (
	"fmt"
	"testing"
)

func TestURL(t *testing.T){
	dups :=0
	all:= make(map[string]int)
	for i:=0;i<200000000;i++{
		a := GetURL()
		if v,ok := all[a]; ok{
			dups++
			fmt.Printf("%5d Duplicates: %s %d %d\n",dups,a,v,i)
		}else{
			t.Error(a)
			t.Fail()
		}
		all[a]=i
		if i < 100 || i%100000==0 {
			fmt.Println(i," ",a)
		}
	}
}