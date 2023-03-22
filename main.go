// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"github.com/pegnet/LXRPow/miner"
)

const LxrPoW_time int64 = 120

func main() {
	c := miner.NewConfig()
	m := new(miner.Miner)
	m.Init(c)
	m.Run()

}
