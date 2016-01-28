package main

import (
        "flag"
        "fmt"
        "github.com/alouca/gosnmp"
        "log"
        "net"
        "strconv"
        "time"
        //"reflect"
)

/*
 fire01 - mappings:
 IF-MIB::ifDescr.4 = STRING: igb3 - DMZ
 IF-MIB::ifDescr.5 = STRING: igb4 - LAN
 IF-MIB::ifDescr.6 = STRING: igb5 - WAN
*/

const (
        snmpagent      = "fire01"
        InBytesDMZOID  = ".1.3.6.1.2.1.31.1.1.1.6.4"  // IF-MIB::ifHCInOctets.4 = Counter64
        InBytesLANOID  = ".1.3.6.1.2.1.31.1.1.1.6.5"  // IF-MIB::ifHCInOctets.5 = Counter64
        InBytesWANOID  = ".1.3.6.1.2.1.31.1.1.1.6.6"  // IF-MIB::ifHCInOctets.6 = Counter64
        OutBytesDMZOID = ".1.3.6.1.2.1.31.1.1.1.10.4" //IF-MIB::ifHCOutOctets.4
        OutBytesLANOID = ".1.3.6.1.2.1.31.1.1.1.10.5" //IF-MIB::ifHCOutOctets.5
        OutBytesWANOID = ".1.3.6.1.2.1.31.1.1.1.10.6" //IF-MIB::ifHCOutOctets.6
)

var (
        Community  string
        Target     string
        cmdTimeout int64 = 5
        OutputAddr string
)

func init() {

        flag.StringVar(&Target, "target", "192.168.1.1", "Address of snmp server. without port.")
        flag.StringVar(&Community, "community", "xxxxxxx", "Community string.")
        flag.StringVar(&OutputAddr, "outputaddr", "localhost:8181", "udp output address")

        flag.Parse()

}

func errchk(err error, msg string) {

        if err != nil {
                log.Printf(msg)
                panic(err)
        }
}

// desc is INtotal.bytes.blabla. (for influx series name)
func GetTraffic(OID string, desc string) (data string) {

        snmp, err := gosnmp.NewGoSNMP(Target, Community, gosnmp.Version2c, cmdTimeout)
        errchk(err, "Could not connect to SNMP target, check target, community")
        snmp.SetTimeout(cmdTimeout)

        var OldVal int64
        var NewVal int64
        ifOldBytes, err := snmp.Get(OID)
        errchk(err, "Could not get OID result")

        for _, v := range ifOldBytes.Variables {

                if OldBytes, err := v.Value.(int64); err {
                        OldVal = OldBytes
                } else {
                        // the above should return a int64
                        fmt.Println("not an int64")
                }

        }
        // sleep for 1 sec then do it again
        fmt.Println("sleeping for 1 second")
        time.Sleep(time.Second * 1)

        ifNewBytes, err := snmp.Get(OID)
        errchk(err, "Could not get OID result")

        for _, v := range ifNewBytes.Variables {

                if NewBytes, err := v.Value.(int64); err {
                        NewVal = NewBytes
                } else {
                        // the above should return a int64
                        fmt.Println("not a int64")
                }

        }
	// print for debug.
        fmt.Println(OldVal)
        fmt.Println(NewVal)
        Diff := NewVal - OldVal
        fmt.Println(Diff)
        // diff is in bytes. convert it into bits.
        Diff = Diff * 8
        // TO DO----- seperate into another function // we will write this data to net to send to heka. -------- // SEND TO HEKA
        // conn, err := net.Dial("udp", "localhost:801"); err = conn.Write(...)
        con, err := net.Dial("udp", OutputAddr)
	// TO DO ----- not sure if i need to check for err here, as its udp and doesnt actually connect.. TO DO -------
        errchk(err, "could not connect to udp endpoint")

        // convert diff into a string. base 10 
        t := strconv.FormatInt(Diff, 10)

        data = snmpagent + desc + ":" + t + "|g"
	// print for debug
        fmt.Fprintf(con, data)
	// close con
        con.Close()

        return data
}

func main() {

        // need to make a for loop here.
        for {
                DMZInTraf := GetTraffic(InBytesDMZOID, "InBytesDMZ")
                fmt.Println(DMZInTraf)

                LANInTraf := GetTraffic(InBytesLANOID, "InBytesLAN")
                fmt.Println(LANInTraf)

                WANInTraf := GetTraffic(InBytesWANOID, "InBytesWAN")
                fmt.Println(WANInTraf)

                DMZOutTraf := GetTraffic(OutBytesDMZOID, "OutBytesDMZ")
                fmt.Println(DMZOutTraf)

                LANOutTraf := GetTraffic(OutBytesLANOID, "OutBytesLAN")
                fmt.Println(LANOutTraf)

                WANOutTraf := GetTraffic(OutBytesWANOID, "OutBytesWAN")
                fmt.Println(WANOutTraf)

                fmt.Println("sleeping for 10 second")
                time.Sleep(time.Second * 10)

        }

}
