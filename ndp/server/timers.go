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
package server

import (
	"github.com/google/gopacket/layers"
	"l3/ndp/config"
	"l3/ndp/debug"
	"time"
)

/*
 *  Stop ReTransmit Timer
 */
func (c *NeighborInfo) StopReTransmitTimer() {
	if c.RetransTimer != nil {
		debug.Logger.Debug("Stopping re-transmit timer for Neighbor", c.IpAddr)
		c.RetransTimer.Stop()
		c.RetransTimer = nil
	}
}

/*
 *  Stop Reachable Timer
 */
func (c *NeighborInfo) StopReachableTimer() {
	if c.ReachableTimer != nil {
		debug.Logger.Debug("Stopping reachable timer for Neighbor", c.IpAddr)
		c.ReachableTimer.Stop()
		c.ReachableTimer = nil
	}
}

/*
 *  Stop Reachable Timer
 */
func (c *NeighborInfo) StopReComputeBaseTimer() {
	if c.RecomputeBaseTimer != nil {
		debug.Logger.Debug("Stopping re-compute timer for Neighbor", c.IpAddr)
		c.RecomputeBaseTimer.Stop()
		c.RecomputeBaseTimer = nil
	}
}

/*
 *  stop delay probe Timer
 */
func (c *NeighborInfo) StopDelayProbeTimer() {
	if c.DelayFirstProbeTimer != nil {
		debug.Logger.Debug("Stopping DelayFirstProbeTimer for Neighbor", c.IpAddr)
		c.DelayFirstProbeTimer.Stop()
		c.DelayFirstProbeTimer = nil
	}
}

/*
 *  stop delay probe Timer
 */
func (intf *Interface) StopRATimer() {
	if intf.raTimer != nil {
		debug.Logger.Debug("Stopping RA Timer for interface:", intf.IntfRef)
		intf.raTimer.Stop()
		intf.raTimer = nil
	}
}

/*
 *  stop delay probe Timer
 */
func (c *NeighborInfo) StopInvalidTimer() {
	if c.InvalidationTimer != nil {
		debug.Logger.Debug("Stopping InvalidationTimer for Neighbor", c.IpAddr)
		c.InvalidationTimer.Stop()
		c.InvalidationTimer = nil
	}
}

/*
 * Delay first probe timer handler
 */
func (c *NeighborInfo) DelayProbe() {
	if c.DelayFirstProbeTimer != nil {
		// we should never come here
		debug.Logger.Debug("Resetting delay probe timer for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr, "to timer:", DELAY_FIRST_PROBE_TIME)
		c.DelayFirstProbeTimer.Reset(time.Duration(DELAY_FIRST_PROBE_TIME) * time.Second)
	} else {
		var DelayProbeExpired_func func()
		DelayProbeExpired_func = func() {
			debug.Logger.Debug("Delay Probe Timer Expired for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr,
				"Sending Probe NS msgs")
			c.ReturnCh <- config.PacketData{
				SendPktType: layers.ICMPv6TypeNeighborSolicitation,
				NeighborIp:  c.IpAddr,
				NeighborMac: c.LinkLayerAddress,
				IfIndex:     c.IfIndex,
			}
		}
		debug.Logger.Debug("Setting Delay Probe timer for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr, "to timer:", DELAY_FIRST_PROBE_TIME)
		c.DelayFirstProbeTimer = time.AfterFunc(time.Duration(DELAY_FIRST_PROBE_TIME)*time.Second,
			DelayProbeExpired_func)
	}
}

/*
 *    Re-Transmit Timer
 */
func (c *NeighborInfo) Timer() {
	// Reset the timer if it is already running when we receive Neighbor Advertisment..
	if c.RetransTimer != nil {
		debug.Logger.Debug("Resetting Re-Transmit timer for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr, "to timer:", c.RetransTimerConfig)
		c.RetransTimer.Reset(time.Duration(c.RetransTimerConfig) * time.Millisecond)
	} else {
		// start the time for the first... provide an after func and move on
		var ReTransmitNeighborSolicitation_func func()
		ReTransmitNeighborSolicitation_func = func() {
			debug.Logger.Debug("Re-Transmit Timer Expired for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr,
				"Sending Neighbor Solicitation")
			c.ReturnCh <- config.PacketData{
				SendPktType: layers.ICMPv6TypeNeighborSolicitation,
				NeighborIp:  c.IpAddr,
				NeighborMac: c.LinkLayerAddress,
				IfIndex:     c.IfIndex,
			}
		}
		debug.Logger.Debug("Setting Re-Transmit timer for ifIndex:", c.IfIndex, "nbrIp:", c.IpAddr, "to timer:", c.RetransTimerConfig)
		c.RetransTimer = time.AfterFunc(time.Duration(c.RetransTimerConfig)*time.Millisecond,
			ReTransmitNeighborSolicitation_func)
	}
}

/*
 *  Start Reachable Timer
 */
func (c *NeighborInfo) RchTimer() {
	// When reachable timer is running or updated we need to stop delay probe timer and re-transmit timer
	// no matter what happens
	c.StopDelayProbeTimer()
	c.StopReTransmitTimer()
	if c.ReachableTimer != nil {
		//debug.Logger.Debug("Re-Setting Reachable Timer for neighbor:", c.IpAddr, "timer:", c.BaseReachableTimer)
		//Reset the timer as we have received an advertisment for the neighbor
		c.ReachableTimer.Reset(time.Duration(c.BaseReachableTimer) * time.Minute)
	} else {
		// This is first time initialization of reachable timer... let set it up
		var ReachableTimer_func func()
		ReachableTimer_func = func() {
			debug.Logger.Debug("Reachable Timer expired for neighbor:", c.IpAddr,
				"initiating unicast NS for ifIndex:", c.IfIndex)
			c.ReturnCh <- config.PacketData{
				SendPktType: layers.ICMPv6TypeNeighborSolicitation,
				NeighborIp:  c.IpAddr,
				NeighborMac: c.LinkLayerAddress,
				IfIndex:     c.IfIndex,
			}
		}
		debug.Logger.Debug("Setting Reachable Timer for neighbor:", c.IpAddr, "timer:", c.BaseReachableTimer)
		c.ReachableTimer = time.AfterFunc(time.Duration(c.BaseReachableTimer)*time.Minute,
			ReachableTimer_func)
	}
}

/*
 *  Re-computing base reachable timer
 */
func (c *NeighborInfo) ReComputeBaseReachableTimer() {
	if c.RecomputeBaseTimer != nil {
		// We need to recompute this timer on RA packets
	} else {
		// set go after function to recompute the time and also restart the timer after that
		var RecomputeBaseTimer_func func()
		RecomputeBaseTimer_func = func() {
			c.BaseReachableTimer = computeBase(c.ReachableTimeConfig)
			c.ReachableTimer.Reset(time.Duration(c.BaseReachableTimer) * time.Minute)
		}
		debug.Logger.Debug("Setting Recompute Timer for neighbor:", c.IpAddr)
		c.RecomputeBaseTimer = time.AfterFunc(time.Duration(RECOMPUTE_BASE_REACHABLE_TIMER)*time.Hour,
			RecomputeBaseTimer_func)
	}
}

/*
 * Router Advertisment Timer: Only timer owned by Interface Object
 */
func (intf *Interface) RAResTransmitTimer() {
	if intf.PcapBase.Tx == nil {
		intf.StopRATimer()
		return
	}
	if intf.raTimer != nil {
		if intf.initialRASend < MAX_INITIAL_RTR_ADVERTISEMENTS {
			intf.raTimer.Reset(time.Duration(MAX_INITIAL_RTR_ADVERT_INTERVAL) * time.Second)
			intf.initialRASend++
		} else {
			//debug.Logger.Debug("Re-Setting ra retransmit timer for intf:", intf.IntfRef, "to:", intf.raRestransmitTime)
			intf.raTimer.Reset(time.Duration(intf.raRestransmitTime) * time.Second)
		}
	} else {
		var raReTransmit_func func()
		raReTransmit_func = func() {
			intf.PktDataCh <- config.PacketData{
				SendPktType: layers.ICMPv6TypeRouterAdvertisement,
				IfIndex:     intf.IfIndex,
			}
		}
		debug.Logger.Debug("Setting ra retransmit timer for intf:", intf.IntfRef, "to:", intf.raRestransmitTime)
		intf.initialRASend = 0
		intf.raTimer = time.AfterFunc(time.Duration(MAX_INITIAL_RTR_ADVERT_INTERVAL)*time.Second,
			raReTransmit_func)
	}
}

/*
 *  invalidation timer received during RA
 */
func (c *NeighborInfo) InValidTimer(lifetime uint16) {
	if c.InvalidationTimer != nil {
		c.InvalidationTimer.Reset(time.Duration(lifetime) * time.Second)
	} else {
		var InvalidationTimer_func func()
		InvalidationTimer_func = func() {
			debug.Logger.Debug("Router Lifetime/Invalidation Timer Expired:", c.IpAddr, "Sending Delete Request")
			// @TODO: Post Delete operation
		}
		debug.Logger.Debug("Setting Router Lifetime/Invalidation Timer", c.IpAddr)
		c.InvalidationTimer = time.AfterFunc(time.Duration(lifetime)*time.Second, InvalidationTimer_func)
	}
}

/*
 *  Update Probe Information
 *	1) Stop Delay Timer if running
 *	2) Stop Re-Transmit Timer if running
 *	3) Update Probes Sent counter to 0
 */
func (c *NeighborInfo) UpdateProbe() {
	debug.Logger.Debug("UpdateProbe info by stopping delay probe timer & re-transmit timer")
	c.StopDelayProbeTimer()
	c.StopReTransmitTimer()
	c.ProbesSent = uint8(0)
}
