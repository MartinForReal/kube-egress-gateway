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

package main

import (
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	type100 "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"

	cniprotocol "github.com/Azure/kube-egress-gateway/pkg/cniprotocol/v1"
	"github.com/Azure/kube-egress-gateway/pkg/consts"
)

const (
	ifName = "eth0"
)

func createCmdArgs(targetNS ns.NetNS) *skel.CmdArgs {
	conf := `{"cniVersion":"1.0.0","excludedCIDRs":["1.2.3.4/32","10.1.0.0/16"],"socketPath":"/tmp/cni_grpc_test.sock","gatewayName":"test","ipam":{"type":"static","addresses":[{"address":"fe80::5/64"},{"address":"10.4.0.5/24"}]},"name":"mynet","type":"kube-egress-cni","prevResult":{"cniVersion":"1.0.0","interfaces":[{"name":"eth0","sandbox":"somepath"}],"ips":[{"interface":0,"address":"10.2.0.1/24"}],"dns":{}}}`
	return &skel.CmdArgs{
		Args:        `IgnoreUnknown=true;K8S_POD_NAMESPACE=testns;K8S_POD_NAME=testpod`,
		ContainerID: "test-container",
		Netns:       targetNS.Path(),
		IfName:      ifName,
		StdinData:   []byte(conf),
	}
}

var _ = Describe("Test kube-egress-cni-ipam operations", func() {
	var originalNS, targetNS ns.NetNS
	var args *skel.CmdArgs
	var ipv4Net, ipv6Net *net.IPNet

	BeforeEach(func() {
		var err error
		originalNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())
		targetNS, err = testutils.NewNS()
		Expect(err).NotTo(HaveOccurred())
		args = createCmdArgs(targetNS)
		ipv4Net, err = netlink.ParseIPNet("10.4.0.5/24")
		Expect(err).NotTo(HaveOccurred())
		ipv6Net, err = netlink.ParseIPNet("fe80::5/64")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("IS_UNIT_TEST_ENV", "true")
	})

	AfterEach(func() {
		os.Setenv("IS_UNIT_TEST_ENV", "")
		Expect(originalNS.Close()).To(Succeed())
		Expect(testutils.UnmountNS(originalNS)).To(Succeed())
		Expect(targetNS.Close()).To(Succeed())
		Expect(testutils.UnmountNS(targetNS)).To(Succeed())
	})

	It("should do nothing and print prevResult when pod does not use any staticGatewayConfiguration", func() {
		grpcTestServer, err := cniprotocol.StartTestServer("/tmp/cni_grpc_test.sock", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(grpcTestServer).NotTo(BeNil())
		defer grpcTestServer.GracefulStop()
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			msg := <-grpcTestServer.Received
			req, ok := msg.(*cniprotocol.PodRetrieveRequest)
			Expect(ok).To(BeTrue())
			Expect(req.GetPodConfig().GetPodNamespace()).To(Equal("testns"))
			Expect(req.GetPodConfig().GetPodName()).To(Equal("testpod"))
			resultType, err := r.GetAsVersion(type100.ImplementedSpecVersion)
			Expect(err).NotTo(HaveOccurred())
			result := resultType.(*type100.Result)
			Expect(len(result.Interfaces)).To(Equal(1))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should configure pod namespace as expected in cmdAdd", func() {
		err := targetNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			Expect(netlink.LinkAdd(&netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: ifName}})).To(Succeed())
			eth0, err := netlink.LinkByName(ifName)
			Expect(err).NotTo(HaveOccurred())
			Expect(netlink.AddrAdd(eth0, &netlink.Addr{IPNet: ipv4Net})).To(Succeed())
			Expect(netlink.AddrAdd(eth0, &netlink.Addr{IPNet: ipv6Net})).To(Succeed())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		grpcTestServer, err := cniprotocol.StartTestServer("/tmp/cni_grpc_test.sock", nil, map[string]string{consts.CNIGatewayAnnotationKey: "test-sgw"})
		Expect(err).NotTo(HaveOccurred())
		Expect(grpcTestServer).NotTo(BeNil())
		defer grpcTestServer.GracefulStop()
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			r, _, err := testutils.CmdAddWithArgs(args, func() error {
				origCNIPath := os.Getenv("CNI_PATH")
				os.Setenv("CNI_PATH", "./testdata") // contains static ipam plugin
				defer os.Setenv("CNI_PATH", origCNIPath)
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			resultType, err := r.GetAsVersion(type100.ImplementedSpecVersion)
			Expect(err).NotTo(HaveOccurred())
			result := resultType.(*type100.Result)
			Expect(len(result.Interfaces)).To(Equal(2))
			Expect(len(result.IPs)).To(Equal(2))

			msg := <-grpcTestServer.Received
			req1, ok := msg.(*cniprotocol.PodRetrieveRequest)
			Expect(ok).To(BeTrue())
			Expect(req1.GetPodConfig().GetPodNamespace()).To(Equal("testns"))
			Expect(req1.GetPodConfig().GetPodName()).To(Equal("testpod"))
			msg = <-grpcTestServer.Received
			req2, ok := msg.(*cniprotocol.NicAddRequest)
			Expect(ok).To(BeTrue())
			Expect(req2.GetPodConfig().GetPodNamespace()).To(Equal("testns"))
			Expect(req2.GetPodConfig().GetPodName()).To(Equal("testpod"))
			Expect(req2.GetGatewayName()).To(Equal("test-sgw"))
			Expect(req2.GetAllowedIp()).To(Equal("10.4.0.5/32"))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not report error in cmdDel", func() {
		grpcTestServer, err := cniprotocol.StartTestServer("/tmp/cni_grpc_test.sock", nil, map[string]string{consts.CNIGatewayAnnotationKey: "test-sgw"})
		Expect(err).NotTo(HaveOccurred())
		Expect(grpcTestServer).NotTo(BeNil())
		defer grpcTestServer.GracefulStop()
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err := testutils.CmdDelWithArgs(args, func() error {
				origCNIPath := os.Getenv("CNI_PATH")
				os.Setenv("CNI_PATH", "./testdata") // contains static ipam plugin
				defer os.Setenv("CNI_PATH", origCNIPath)
				return cmdDel(args)
			})
			Expect(err).NotTo(HaveOccurred())

			msg := <-grpcTestServer.Received
			req, ok := msg.(*cniprotocol.NicDelRequest)
			Expect(ok).To(BeTrue())
			Expect(req.GetPodConfig().GetPodNamespace()).To(Equal("testns"))
			Expect(req.GetPodConfig().GetPodName()).To(Equal("testpod"))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not report error in cmdCheck", func() {
		err := originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err := testutils.CmdCheckWithArgs(args, func() error {
				return cmdCheck(args)
			})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
})