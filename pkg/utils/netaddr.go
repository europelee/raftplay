package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	addrSepFlag     = ":"
	addrListSepFlag = ","
)

//NetAddr <IP, Port>
type NetAddr struct {
	IP   string
	Port uint16
}

//NetAddrList []NetAddr
type NetAddrList struct {
	NetAddrs []NetAddr
}

//Set parse <IP>:<Port>
func (p *NetAddr) Set(s string) error {
	valSlice := strings.Split(s, addrSepFlag)
	if len(valSlice) != 2 {
		return errors.New("invalid netAddr")
	}
	p.IP = valSlice[0]
	port, err := strconv.ParseUint(valSlice[1], 10, 16)
	if err != nil {
		return errors.New("invalid netAddr")
	}
	p.Port = uint16(port)
	return nil
}

func (p *NetAddr) String() string {
	return fmt.Sprintf("%v%v%v", p.IP, addrSepFlag, p.Port)
}

//Set parse <netaddr>,<netaddr>
func (nl *NetAddrList) Set(s string) error {
	varSlice := strings.Split(s, addrListSepFlag)
	for _, netAddrStr := range varSlice {
		var addrInst NetAddr
		err := addrInst.Set(netAddrStr)
		if err != nil {
			return err
		}
		nl.NetAddrs = append(nl.NetAddrs, addrInst)
	}
	return nil
}

func (nl *NetAddrList) String() string {
	return fmt.Sprintf("%+v", *nl)
}
