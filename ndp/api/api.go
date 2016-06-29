//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//
package api

import (
	"l3/ndp/config"
	"l3/ndp/server"
	"sync"
)

var ndpApi *NDPApiLayer = nil
var once sync.Once

type NDPApiLayer struct {
	server *server.NDPServer
}

/*  Singleton instance should be accessible only within api
 */
func getApiInstance() *NDPApiLayer {
	once.Do(func() {
		ndpApi = &NDPApiLayer{}
	})
	return ndpApi
}

func Init(svr *server.NDPServer) {
	ndpApi = getApiInstance()
	ndpApi.server = svr
}

func SendL2PortNotification(ifIndex int32, state string) {

}

func SendL3PortNotification(ifIndex int32, state string) {

}

func SendVlanNotification(vlanId uint16, vlanName, state string) {

}

func SendIPv6Notfication(ifIndex int32, ipaddr, msgType string) {
	ndpApi.server.Ipv6Ch <- &config.IPv6IntfInfo{
		IfIndex:   ifIndex,
		IpAddr:    ipaddr,
		OperState: msgType,
	}
}
