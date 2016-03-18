package vrrpRpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"io/ioutil"
	"l3/vrrp/server"
	"strconv"
	"utils/logging"
	"vrrpd"
)

type VrrpHandler struct {
	server *vrrpServer.VrrpServer
	logger *logging.Writer
}
type VrrpClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

const (
	VRRP_RPC_NO_PORT = "could not find port and hence not starting rpc"
)

func (h *VrrpHandler) CreateVrrpIntf(config *vrrpd.VrrpIntf) (r bool, err error) {
	h.logger.Info(fmt.Sprintln("VRRP: Interface config create for ifindex ",
		config.IfIndex))
	if config.VRID == 0 {
		h.logger.Info("VRRP: Invalid VRID")
		return false, errors.New(vrrpServer.VRRP_INVALID_VRID)
	}

	err = h.server.VrrpValidateIntfConfig(config.IfIndex)
	if err != nil {
		return false, err
	}
	h.server.VrrpCreateIntfConfigCh <- *config
	return true, err
}
func (h *VrrpHandler) UpdateVrrpIntf(origconfig *vrrpd.VrrpIntf,
	newconfig *vrrpd.VrrpIntf, attrset []bool) (r bool, err error) {
	return true, nil
}

func (h *VrrpHandler) DeleteVrrpIntf(config *vrrpd.VrrpIntf) (r bool, err error) {
	h.server.VrrpDeleteIntfConfigCh <- *config
	return true, nil
}

func (h *VrrpHandler) convertVrrpEntryToThriftEntry(state vrrpd.VrrpIntfState) *vrrpd.VrrpIntfState {
	entry := vrrpd.NewVrrpIntfState()
	entry.VirtualRouterMACAddress = state.VirtualRouterMACAddress
	entry.PreemptMode = bool(state.PreemptMode)
	entry.AdvertisementInterval = int32(state.AdvertisementInterval)
	entry.VRID = int32(state.VRID)
	entry.Priority = int32(state.Priority)
	entry.SkewTime = int32(state.SkewTime)
	entry.VirtualIPv4Addr = state.VirtualIPv4Addr
	entry.IfIndex = int32(state.IfIndex)
	entry.MasterDownTimer = int32(state.MasterDownTimer)
	entry.IntfIpAddr = state.IntfIpAddr
	entry.VrrpState = state.VrrpState
	return entry
}

func (h *VrrpHandler) GetBulkVrrpIntfState(fromIndex vrrpd.Int,
	count vrrpd.Int) (*vrrpd.VrrpIntfStateGetInfo, error) {
	nextIdx, currCount, vrrpIntfStateEntries := h.server.VrrpGetBulkVrrpIntfStates(
		int(fromIndex), int(count))
	if vrrpIntfStateEntries == nil {
		return nil, errors.New("Interface Slice is not initialized")
	}
	vrrpEntryResponse := make([]*vrrpd.VrrpIntfState, len(vrrpIntfStateEntries))
	for idx, item := range vrrpIntfStateEntries {
		vrrpEntryResponse[idx] = h.convertVrrpEntryToThriftEntry(item)
	}
	intfEntryBulk := vrrpd.NewVrrpIntfStateGetInfo()
	intfEntryBulk.VrrpIntfStateList = vrrpEntryResponse
	intfEntryBulk.StartIdx = fromIndex
	intfEntryBulk.EndIdx = vrrpd.Int(nextIdx)
	intfEntryBulk.Count = vrrpd.Int(currCount)
	intfEntryBulk.More = (nextIdx != 0)
	return intfEntryBulk, nil
}

func VrrpNewHandler(vrrpSvr *vrrpServer.VrrpServer, logger *logging.Writer) *VrrpHandler {
	hdl := new(VrrpHandler)
	hdl.server = vrrpSvr
	hdl.logger = logger
	return hdl
}

func VrrpRpcGetClient(logger *logging.Writer, fileName string, process string) (*VrrpClientJson, error) {
	var allClients []VrrpClientJson

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Err(fmt.Sprintf("Failed to open VRRPd config file:%s, err:%s", fileName, err))
		return nil, err
	}

	json.Unmarshal(data, &allClients)
	for _, client := range allClients {
		if client.Name == process {
			return &client, nil
		}
	}

	logger.Err(fmt.Sprintf("Did not find port for %s in config file:%s", process, fileName))
	return nil, errors.New(VRRP_RPC_NO_PORT)

}

func StartServer(log *logging.Writer, handler *VrrpHandler, paramsDir string) error {
	logger := log
	fileName := paramsDir

	if fileName[len(fileName)-1] != '/' {
		fileName = fileName + "/"
	}
	fileName = fileName + "clients.json"

	clientJson, err := VrrpRpcGetClient(logger, fileName, "vrrpd")
	if err != nil || clientJson == nil {
		return err
	}
	logger.Info(fmt.Sprintln("Got Client Info for", clientJson.Name, " port",
		clientJson.Port))
	// create processor, transport and protocol for server
	processor := vrrpd.NewVRRPDServicesProcessor(handler)
	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transport, err := thrift.NewTServerSocket("localhost:" + strconv.Itoa(clientJson.Port))
	if err != nil {
		logger.Info(fmt.Sprintln("StartServer: NewTServerSocket "+
			"failed with error:", err))
		return err
	}
	server := thrift.NewTSimpleServer4(processor, transport,
		transportFactory, protocolFactory)
	err = server.Serve()
	if err != nil {
		logger.Err(fmt.Sprintln("Failed to start the listener, err:", err))
		return err
	}
	return nil
}
