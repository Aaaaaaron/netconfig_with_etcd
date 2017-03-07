package main

//get data from etcd watch.
func main() {
	links := GetLinkDetailsInJSON()
	PutIntoEtcd(links)
	MonitorEtcd()
	MonitorLinks()
	//ProcessUserConfig() 这个用户操作etcd 不是在这层的
}
func MonitorEtcd() {
	//if changes notify to change link setting use netlink
}
func ProcessUserConfig() {
	switch req {
	case listLink:
		Listlink()
	case addLink:
		Addlink()
	case delLink:
		DelLink()
	case updateLink:
		UpdateLink()
	}
}
