package main

import (
	//	"encoding/binary"
	"flag"
	"log"
	"net"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/unix"
)

type PacketMeta struct {
	dstIP   net.IP
	dstMAC  net.HardwareAddr
	dstPort uint16
	file    string
	iface   string
	payload int
	srcIP   net.IP
	srcMAC  net.HardwareAddr
	srcPort uint16
}

type Options struct {
	dstIP   *string
	dstMAC  *string
	dstPort *int
	file    *string
	iface   *string
	payload *int
	srcIP   *string
	srcMAC  *string
	srcPort *int
}

func CheckPort(direction string, port int) uint16 {
	if port < 0 {
		log.Fatalf("%s port must be > 0\n", direction)
	}

	if port > 65535 {
		log.Fatalf("%s port must be < 65535\n", direction)
	}

	return uint16(port)
}

func ValidateOptions(o Options) *PacketMeta {
	var p PacketMeta
	var err error

	p.srcMAC, err = net.ParseMAC(*o.srcMAC)
	if err != nil {
		log.Fatalf("Src MAC Address: %v", err)
	}

	p.dstMAC, err = net.ParseMAC(*o.dstMAC)
	if err != nil {
		log.Fatalf("Dst MAC Address: %v", err)
	}

	p.srcIP = net.ParseIP(*o.srcIP)
	if p.srcIP == nil {
		log.Fatalf("Src IP address: %s", *o.srcIP)
	}

	p.dstIP = net.ParseIP(*o.dstIP)
	if p.dstIP == nil {
		log.Fatalf("Dst IP address: %s", *o.dstIP)
	}

	p.srcPort = CheckPort("src", *o.srcPort)
	p.dstPort = CheckPort("dst", *o.dstPort)

	p.file = *o.file
	p.iface = *o.iface
	p.payload = *o.payload

	return &p
}

func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

func MakePacket(pm PacketMeta) []byte {
	buf := gopacket.NewSerializeBuffer()
	var packet []gopacket.SerializableLayer

	ethLayer := &layers.Ethernet{
		SrcMAC:       pm.srcMAC,
		DstMAC:       pm.dstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	packet = append(packet, ethLayer)

	ipLayer := &layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    pm.srcIP,
		DstIP:    pm.dstIP,
		Protocol: layers.IPProtocolUDP,
	}
	packet = append(packet, ipLayer)

	udpLayer := &layers.UDP{
		SrcPort: layers.UDPPort(pm.srcPort),
		DstPort: layers.UDPPort(pm.dstPort),
	}
	udpLayer.SetNetworkLayerForChecksum(ipLayer) // Important for checksum calculation
	packet = append(packet, udpLayer)

	payload := make([]byte, pm.payload)
	// Optionally, fill the payload with data
	packet = append(packet, gopacket.Payload(payload))

	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}, packet...); err != nil {
		log.Fatalf("error serializing packet: %v", err)
	}

	return buf.Bytes()
}

func main() {
	var opt Options
	opt.dstIP = flag.String("ip.dst", "100.100.100.100", "Specify the IPv4 destination address")
	opt.dstMAC = flag.String("mac.dst", "dc:a6:32:9d:c1:6e", "Specify the destination Ethernet MAC address")
	opt.dstPort = flag.Int("port.dst", 514, "Specify the destination port")
	opt.iface = flag.String("iface", "eth0", "Specify the network interface to use")
	opt.file = flag.String("file", "", "Specify the file to use as packet content")
	opt.payload = flag.Int("payload", 128, "Specify the payload size")
	opt.srcIP = flag.String("ip.src", "192.168.245.51", "Specify the IPv4 source address")
	opt.srcMAC = flag.String("mac.src", "dc:a6:32:9d:c0:f9", "Specify the source Ethernet MAC address")
	opt.srcPort = flag.Int("port.src", 6000, "Specify the source port")
	flag.Parse()

	pm := ValidateOptions(opt)

	//htons := make([]byte, 2)
	//binary.BigEndian.PutUint16(htons, syscall.ETH_P_IP)
	//log.Printf("%d %T\n", syscall.ETH_P_IP, syscall.ETH_P_IP)
	//log.Println(htons)
	//fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_IP)))
	fd, err := syscall.Socket(unix.AF_PACKET, syscall.SOCK_RAW, int(htons(unix.ETH_P_IP)))
	if err != nil {
		log.Fatal("failed to create socket: %v", err)
	}
	defer syscall.Close(fd)

	ifi, err := net.InterfaceByName(pm.iface)
	if err != nil {
		log.Fatalf("failed to get interface %s: %v", pm.iface, err)
	}

	addr := &unix.SockaddrLinklayer{
		Protocol: htons(unix.ETH_P_IP),
		Ifindex:  ifi.Index,
	}

	packet := MakePacket(*pm)
	err = unix.Sendto(fd, packet, 0, addr)
	if err != nil {
		log.Println(err)
	}
}
