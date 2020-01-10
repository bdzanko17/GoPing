package main

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"time"
)

var (
	device       string = "wlp6s0"
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = 1 * time.Second
	handle       *pcap.Handle
	TSval        uint32 = 0
	TSerc        uint32 = 0
	source_ip    net.IP
	dest_ip      net.IP
	source_port  string
	dest_port    string
	srcstr       string
	dststr       string
)

type Flowrecord struct {
	last_time float64
	flowname  string
}

func main() {
	Flows := make(map[string]bool)
	TimeFlow := make(map[string]int64)
	var vrijeme int64
	var br_paketa int
	var br_flow int

	// Open device
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	var filter string = "tcp and port 80"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		br_paketa++

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			source_ip = ip.SrcIP
			dest_ip = ip.DstIP
			fmt.Println(source_ip)
			fmt.Println(dest_ip)
		}
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if len(tcp.Options) >= 3 {

				TSval = binary.BigEndian.Uint32(tcp.Options[2].OptionData[:4])
				TSerc = binary.BigEndian.Uint32(tcp.Options[2].OptionData[4:8])
			}

			source_port = tcp.SrcPort.String()
			dest_port = tcp.DstPort.String()

		}
		srcstr = source_ip.String() + ":" + source_port
		dststr = dest_ip.String() + ":" + dest_port

		var fstr string
		fstr = srcstr + dststr

		vrijeme = time.Now().UnixNano()
		_, ok := Flows[dststr+srcstr]
		if ok {
			Flows[fstr] = true
			Flows[dststr+srcstr] = true
			TimeFlow[dststr+srcstr] = vrijeme
			fmt.Println("RTT: ", (TimeFlow[dststr+srcstr]-TimeFlow[fstr])/1000, "ms")
			delete(Flows, fstr)
			br_flow++
		}

		Flows[fstr] = false
		TimeFlow[fstr] = vrijeme

		fmt.Println(dststr + srcstr)
	}
	//

}
