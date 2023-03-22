/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/
package routes

import (
	"net"
	"testing"

	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"

	"github.com/Azure/kube-egress-gateway/pkg/iptableswrapper/mockiptableswrapper"
	"github.com/Azure/kube-egress-gateway/pkg/netlinkwrapper/mocknetlinkwrapper"
	"github.com/golang/mock/gomock"
	"github.com/vishvananda/netlink"
)

func TestSetPodRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mnl := mocknetlinkwrapper.NewMockInterface(ctrl)
	mipt := mockiptableswrapper.NewMockInterface(ctrl)
	mtable := mockiptableswrapper.NewMockIpTables(ctrl)
	routesRunner = runner{
		netlink:  mnl,
		iptables: mipt,
	}

	eth0 := &netlink.Device{LinkAttrs: netlink.LinkAttrs{Name: "eth0", Index: 1}}
	wg0 := &netlink.Device{LinkAttrs: netlink.LinkAttrs{Name: "wg0", Index: 2}}
	defaultGw := net.IPv4(10, 244, 0, 1)
	existingRoutes := []netlink.Route{
		// default route
		{
			Family:    nl.FAMILY_V4,
			Gw:        defaultGw,
			LinkIndex: 1,
		},
		{
			Dst:       &net.IPNet{IP: net.IPv4(10, 244, 0, 0), Mask: net.CIDRMask(24, 32)},
			LinkIndex: 1,
		},
	}
	_, net1, _ := net.ParseCIDR("1.2.3.4/32")
	_, net2, _ := net.ParseCIDR("172.17.0.4/16")
	_, dnet, _ := net.ParseCIDR("0.0.0.0/0")
	rule := netlink.NewRule()
	rule.Mark = 8738
	rule.Table = 8738
	defaultRoute := netlink.Route{
		Family:    nl.FAMILY_V4,
		Gw:        defaultGw,
		LinkIndex: 1,
		Table:     8738,
	}
	gomock.InOrder(
		// retrieve eth0 link
		mnl.EXPECT().LinkByName("eth0").Return(eth0, nil),
		// get existing routes
		mnl.EXPECT().RouteList(eth0, netlink.FAMILY_ALL).Return(existingRoutes, nil),
		// delete existing routes
		mnl.EXPECT().RouteDel(&existingRoutes[0]).Return(nil),
		mnl.EXPECT().RouteDel(&existingRoutes[1]).Return(nil),
		// add route to default gateway via eth0
		mnl.EXPECT().RouteReplace(&netlink.Route{
			Dst:       &net.IPNet{IP: defaultGw, Mask: net.CIDRMask(32, 32)},
			LinkIndex: 1,
			Scope:     netlink.SCOPE_LINK,
		}).Return(nil),
		// add routes to exceptional CIDRs via eth0
		mnl.EXPECT().RouteReplace(&netlink.Route{
			Dst:       net1,
			Gw:        defaultGw,
			LinkIndex: 1,
			Protocol:  unix.RTPROT_STATIC,
		}).Return(nil),
		mnl.EXPECT().RouteReplace(&netlink.Route{
			Dst:       net2,
			Gw:        defaultGw,
			LinkIndex: 1,
			Protocol:  unix.RTPROT_STATIC,
		}).Return(nil),
		// retrieve wg0 link
		mnl.EXPECT().LinkByName("wg0").Return(wg0, nil),
		// add default route via wg0
		mnl.EXPECT().RouteReplace(&netlink.Route{
			Dst: dnet,
			Gw:  nil,
			Via: &netlink.Via{
				Addr:       net.ParseIP("fe80::1"),
				AddrFamily: nl.FAMILY_V6,
			},
			LinkIndex: 2,
			Scope:     netlink.SCOPE_UNIVERSE,
			Family:    nl.FAMILY_V4,
		}),
		// add iptables rules
		mipt.EXPECT().New().Return(mtable, nil),
		mtable.EXPECT().AppendUnique("mangle", "PREROUTING", "-i", "eth0", "-j", "MARK", "--set-mark", "8738").Return(nil),
		mtable.EXPECT().AppendUnique("mangle", "PREROUTING", "-j", "CONNMARK", "--save-mark").Return(nil),
		mtable.EXPECT().AppendUnique("mangle", "OUTPUT", "-m", "connmark", "--mark", "8738", "-j", "CONNMARK", "--restore-mark").Return(nil),
		mnl.EXPECT().RuleAdd(rule).Return(nil),
		// add route in 8738 table
		mnl.EXPECT().RouteReplace(&defaultRoute).Return(nil),
	)
	err := SetPodRoutes("wg0", []string{"1.2.3.4/32", "172.17.0.4/16"})
	if err != nil {
		t.Fatalf("SetPodRoutes returns unexpected error: %v", err)
	}
}