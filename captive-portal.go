package main

import (
  "os/exec"
  "net"
  "log"
  "net/http"
  "html/template"
  "strings"
  "fmt"
)

const PortalPort = "80"
const IFace = "wlan0"
const SshDeviceIp = "192.168.2.2"

func main() {
  // 指定したinterfaceのipを取得
  portalIp, err := getPrivateIp(IFace)
  if err != nil {
    panic(err.Error())
  }

  // iptablesを初期化
  err = InitTables(portalIp, IFace)
  if err != nil {
    panic(err.Error())
  }

  // captive portalのwebページを起動
  http.Handle("/static", http.StripPrefix("/static", http.FileServer(http.Dir("static/"))))
  http.HandleFunc("/", handleRegister)
  http.HandleFunc("/approve", handleApprove)
  http.ListenAndServe(":" + PortalPort, nil)
}

func getPrivateIp(ifaceName string) (string, error) {
  iface, err := net.InterfaceByName(ifaceName)
  if err != nil {
    return "", err
  }
  addrs, err := iface.Addrs()
  if err != nil {
    return "", err
  }
  addressWithCidr := addrs[0].String()
  ip, _, err := net.ParseCIDR(addressWithCidr)
  if err != nil {
    return "", err
  }
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
  // 許可していないFORWARDトラフィックはブロックする
  if err := exec.Command("iptables", "-P", "FORWARD", "DROP").Run(); err != nil {
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
  if err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE").Run(); err != nil {
    fmt.Println("hoge")
    return err
  }
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", "eth0", "-o", ifaceName, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  if err := exec.Command("iptables", "-A", "FORWARD", "-i", ifaceName, "-o", "eth0", "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  // HTTPでのトラフィックをCaptivePortalのwebサーバーへリダイレクトする
  portalServer := portalIp + ":" + PortalPort
  if err := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-i", ifaceName, "-p", "tcp", "--dport", PortalPort, "-j", "DNAT", "--to-destination", portalServer).Run(); err != nil {
    return err
  }
  return nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
  t, err := template.ParseFiles("views/register_page.html")
  if err != nil {
    log.Fatalf("template error: %v", err)
  }
  if err := t.Execute(w, nil); err != nil {
    log.Fatalf("failed to execute template: %v", err)
  }
}

func handleApprove(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  approved := strings.Compare(r.FormValue("isApproved"), "on") == 0
  if approved {
    clientIp, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
      log.Fatalf("failed to parse ip of %v", r.RemoteAddr)
      return
    }
    if err := allowTrafic(clientIp); err != nil {
      log.Fatalf("failed to allow traffic from %v. error: %v", clientIp, err)
      return
    }
  }
  t, err := template.ParseFiles("views/connected.html")
  if err != nil {
    log.Fatalf("template error: %v", err)
  }
  if err := t.Execute(w, nil); err != nil {
    log.Fatalf("failed to execute template: %v", err)
  }
}

func allowTrafic(ip string) error {
  if err := exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "1", "-s", ip, "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  if err := exec.Command("iptables", "-I", "FORWARD", "-s", ip, "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  if err := exec.Command("iptables", "-I", "INPUT", "-s", ip, "-j", "ACCEPT").Run(); err != nil {
    return err
  }
  return nil
}
