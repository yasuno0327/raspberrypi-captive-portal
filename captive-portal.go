package main

import (
  "os/exec"
  "net"
  "fmt"
)

const PortalPort = "80"
const IFace = "wlan0"
const SshDeviceIp = "192.168.2.2"

func main() {
  // err := exec.Command("iptables", "-A", "FORWARD", "-i", )
  portalIp, err := getPrivateIp(IFace)
  if err != nil {
    panic(err.Error())
  }

  err = InitTables(portalIp, IFace)
  if err != nil {
    panic(err.Error())
  }

}

func getPrivateIp(ifaceName string) (string, error) {
  iface, err := net.InterfaceByName(ifaceName)
  addrs, err := iface.Addrs()
  addressWithCidr := addrs[0].String()
  ip, _, err := net.ParseCIDR(addressWithCidr)
  return ip.To4().String(), err
}

func InitTables(portalIp string, ifaceName string) error {
  // nat, filterテーブルのルールをクリア
  err := exec.Command("iptables", "-F").Run()
  err = exec.Command("iptables", "-t", "nat", "-F").Run()
  // ラズパイへのsshを許可
  err = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", "ssh", "-s", SshDeviceIp, "-j", "ACCEPT").Run()
  // TCPでのDNSリクエストを許可する
  err = exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "tcp", "--dport", "53", "-j", "ACCEPT").Run()
  // UDPでのDNSリクエストを許可する
  err = exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "udp", "--dport", "53", "-j", "ACCEPT").Run()
  // Captive Portalのwebサーバーへのアクセスを許可する
  err = exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "tcp", "--dport", PortalPort, "-d", portalIp, "-j", "ACCEPT").Run()
  // その他のトラフィックはブロック
  err = exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-j", "DROP").Run()
  // HTTPでのトラフィックをCaptivePortalのwebサーバーへリダイレクトする
  portalServer := portalIp + ":" + PortalPort
  err = exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-i", ifaceName, "-p", "tcp", "--dport", PortalPort, "-j", "DNAT", "--to-destination", portalServer).Run()
  return err
}
