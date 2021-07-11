package main

import (
  "os/exec"
  "net"
)

const PortalPort = "80"
const IFace = "wlan0"
const SshDeviceIp = "192.168.2.2"

func main() {
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
  if err := exec.Command("iptables", "-F").Run(); err != nil {
    return err
  }
  if err := exec.Command("iptables", "-t", "nat", "-F").Run(); err != nil {
    return err
  }
  // ラズパイへのsshを許可
  if err := exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", "ssh", "-s", SshDeviceIp, "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  // TCPでのDNSリクエストを許可する
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "tcp", "--dport", "53", "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  // UDPでのDNSリクエストを許可する
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "udp", "--dport", "53", "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  // Captive Portalのwebサーバーへのアクセスを許可する
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-p", "tcp", "--dport", PortalPort, "-d", portalIp, "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  // その他のトラフィックはブロック
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-j", "DROP").Run(); err != nil {
    return err
  }
  // HTTPでのトラフィックをCaptivePortalのwebサーバーへリダイレクトする
  portalServer := portalIp + ":" + PortalPort
  if err := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-i", ifaceName, "-p", "tcp", "--dport", PortalPort, "-j", "DNAT", "--to-destination", portalServer).Run(); err != nil {
    return err
  }
  return nil
}
