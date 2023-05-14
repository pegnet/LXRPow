// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package pow

import (
	"fmt"
	"testing"
	"time"
)

func TestLxrPow_Init(t *testing.T) {
	for i:= 8; i<31; i++ {
		p := new(LxrPow)
 		p.Init(100,i,6)
		
		var Sums[256] int
		for _,b := range p.ByteMap {
			Sums[b]++
		}
		v :=0
		for _,cnt := range Sums {
			if v != cnt {
				v =  cnt
				fmt.Printf("ByteMap size %16d Bits %2d %02x\n",len(p.ByteMap),i,cnt)
			}
		}
		fmt.Println()
		time.Sleep(0)
	}
}
