// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main



	

const LxrPoW_time int64 = 120

func main() {
	c:=NewConfig()
	c.Init()
	m := new(Miner)
	m.Init(c)
}


	