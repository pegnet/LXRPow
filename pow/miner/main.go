// Copyright (c) of parts are held by the various contributors
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pegnet/LXRPow/pow"
	"github.com/pegnet/LXRPow/pow/miner/hashing"
)

const LxrPoW_time int64 = 120

func main() {
	c := NewConfig()
	
	hashing.NewHashers(c.Instances,c.Seed,c.lx)
}


	