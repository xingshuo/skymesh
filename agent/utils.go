package skymesh

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var defaultIP string = "0.0.0.0"

func GetInternalIP() string { //本机内网ip,取第一个
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return defaultIP
	}
	for _, address := range addrs {
		ipnet, ok := address.(*net.IPNet)
		if !ok {
			continue
		}
		if ipnet.IP.IsLoopback() || ipnet.IP.IsLinkLocalMulticast() || ipnet.IP.IsLinkLocalUnicast() {
			continue
		}
		if ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return defaultIP
}

func GetExternalIP() string { //本机外网ip
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return defaultIP
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return defaultIP
	}
	return string(content)
}

func inet_addr(ipaddr string) uint32 {
	var (
		ip                 = strings.Split(ipaddr, ".")
		ip1, ip2, ip3, ip4 uint64
		ret                uint32
	)
	ip1, _ = strconv.ParseUint(ip[0], 10, 8)
	ip2, _ = strconv.ParseUint(ip[1], 10, 8)
	ip3, _ = strconv.ParseUint(ip[2], 10, 8)
	ip4, _ = strconv.ParseUint(ip[3], 10, 8)
	ret = uint32(ip4)<<24 + uint32(ip3)<<16 + uint32(ip2)<<8 + uint32(ip1)
	return ret
}

var skymeshIP string = ""

func GetSkymeshIP() string { //优先取外网ip, 取不到再取内网ip
	if skymeshIP != "" {
		return skymeshIP
	}
	ip := GetExternalIP() //http请求查询, 会阻塞. 只查一次, 本地缓存
	if ip != defaultIP {
		skymeshIP = ip
		return ip
	}
	return GetInternalIP()
}

//instanceID生成规则:(ip << 32)|Hash(pwd|service_name|serviceID|port)
func MakeServiceHandle(serverAddr string, serviceName string, serviceID uint64) uint64 {
	var (
		ip   string
		port string
	)
	s := strings.Split(serverAddr, ":")
	if len(s) != 2 {
		ip = GetSkymeshIP()
		port = "0"
	} else {
		ip = s[0]
		if net.ParseIP(ip) == nil { //填写的域名??
			ip = GetSkymeshIP()
		}
		port = s[1]
	}
	pwd, _ := os.Getwd()
	low32 := fmt.Sprintf("%s|%s|%d|%s", pwd, serviceName, serviceID, port)
	h := fnv.New32a()
	h.Write([]byte(low32))
	hashID := h.Sum32()
	instID := uint64(inet_addr(ip))
	instID = (instID << 32) | uint64(hashID)
	return instID
}

const skymeshSchema = "skymesh"

//skymesh url 格式地址: [skymesh://]ServiceName/service_id
//ServiceName fmt: appid.env_name.service_name
//注意: skymeshServer.Register时使用, service_id不能省略
func SkymeshUrl2Addr(url string) (*Addr, error) {
	rawUrl := StripSkymeshUrlPrefix(url)
	urlItems := strings.Split(rawUrl, "/")
	if len(urlItems) != 2 {
		return nil, fmt.Errorf("skymesh url shoule be service_name/service_id")
	}
	serviceId, _ := strconv.ParseUint(urlItems[1], 10, 64)
	return &Addr{ServiceName: urlItems[0], ServiceId: serviceId}, nil
}

func SkymeshAddr2Url(addr *Addr, withPrefix bool) string {
	prefix := ""
	if withPrefix {
		prefix = fmt.Sprintf("%s://", skymeshSchema)
	}
	return prefix + fmt.Sprintf("%s/%v", addr.ServiceName, addr.ServiceId)
}

func StripSkymeshUrlPrefix(url string) string {
	prefix := fmt.Sprintf("%s://", skymeshSchema)
	if !strings.HasPrefix(url, prefix) {
		return url
	}
	return url[len(prefix):]
}

func GetSkymeshSchema() string {
	return skymeshSchema
}

//skymesh url fmt: [skymesh://]appid.env_name.service_name[/service_id]
func ParseSkymeshUrl(url string) (appID string, envName string, svcName string, svcInstID uint64, err error) {
	rawUrl := StripSkymeshUrlPrefix(url)
	urlItems := strings.Split(rawUrl, "/")
	fullSvcName := rawUrl
	svcInstID = 0
	if len(urlItems) == 2 { //url include service_id
		fullSvcName = urlItems[0]
		svcInstID,_ = strconv.ParseUint(urlItems[1], 10, 64)
	}
	s := strings.Split(fullSvcName, ".")
	if len(s) != 3 {
		err = fmt.Errorf("skymesh ServiceName fmt err, appid.env_name.service_name expected")
		return
	}
	appID = s[0]
	envName = s[1]
	svcName = s[2]
	return
}

func MakeSkymeshUrl(fullSvcName string, instID uint64) string {
	return fmt.Sprintf("%s/%v",fullSvcName, instID)
}
