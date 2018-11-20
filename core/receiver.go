package core

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func ReceivePeerOperations(conn net.Conn, reducer *Reduce) {
	defer conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)

	for {
		// Read the incoming connection into the buffer.
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		sBuf := string(buf)
		var o PeerOperation
		o.OpType = string(buf[0])
		o.Character = string(buf[1])
		vcStart := strings.Index(sBuf[2:], " ") + 1
		if vcStart <= 0 {
			fmt.Println("Error parsing peer operation")
			return
		}

		if o.Position, err = strconv.Atoi(string(buf[2 : vcStart-1])); err != nil {
			fmt.Println("Error: could not parse position int", err.Error())
		}

		if vClock, err := ParseVectorClock(string(buf[vcStart:n])); err != nil {
			fmt.Println("Error: could not parse vector clock", err.Error())
		} else {
			o.VClock = *vClock
		}

		// send operation to algorithm to be processed
		reducer.PeerPropose(o)
	}
}

func IsPeer(conn net.Conn) bool {
	buf := make([]byte, 256)

	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		os.Exit(1)
	}

	return string(buf[:n]) == "peer"
}
