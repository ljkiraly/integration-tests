// Code generated by gotestmd DO NOT EDIT.
package remotevlan

import (
	"github.com/stretchr/testify/suite"

	"github.com/ljkiraly/integration-tests/extensions/base"
	"github.com/ljkiraly/integration-tests/suites/remotevlan/rvlanovs"
	"github.com/ljkiraly/integration-tests/suites/remotevlan/rvlanvpp"
	"github.com/ljkiraly/integration-tests/suites/spire"
)

type Suite struct {
	base.Suite
	spireSuite    spire.Suite
	rvlanovsSuite rvlanovs.Suite
	rvlanvppSuite rvlanvpp.Suite
}

func (s *Suite) SetupSuite() {
	parents := []interface{}{&s.Suite, &s.spireSuite}
	for _, p := range parents {
		if v, ok := p.(suite.TestingSuite); ok {
			v.SetT(s.T())
		}
		if v, ok := p.(suite.SetupAllSuite); ok {
			v.SetupSuite()
		}
	}
	r := s.Runner("../nsm-deployments-k8s/examples/remotevlan")
	s.T().Cleanup(func() {
		r.Run(`WH=$(kubectl get pods -l app=admission-webhook-k8s -n nsm-system --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')` + "\n" + `kubectl delete mutatingwebhookconfiguration ${WH}` + "\n" + `kubectl delete ns nsm-system`)
		r.Run(`docker network disconnect bridge-2 kind-worker` + "\n" + `docker network disconnect bridge-2 kind-worker2` + "\n" + `docker network rm bridge-2` + "\n" + `docker exec kind-worker ip link del ext_net1` + "\n" + `docker exec kind-worker2 ip link del ext_net1` + "\n" + `true`)
	})
	r.Run(`docker network create bridge-2` + "\n" + `docker network connect bridge-2 kind-worker` + "\n" + `docker network connect bridge-2 kind-worker2`)
	r.Run(`MACS=($(docker inspect --format='{{range .Containers}}{{.MacAddress}}{{"\n"}}{{end}}' bridge-2))` + "\n" + `ifw1=$(docker exec kind-worker ip -o link | grep ${MACS[@]/#/-e } | cut -f1 -d"@" | cut -f2 -d" ")` + "\n" + `ifw2=$(docker exec kind-worker2 ip -o link | grep ${MACS[@]/#/-e } | cut -f1 -d"@" | cut -f2 -d" ")` + "\n" + `` + "\n" + `(docker exec kind-worker ip link set $ifw1 down &&` + "\n" + `docker exec kind-worker ip link set $ifw1 name ext_net1 &&` + "\n" + `docker exec kind-worker ip link set ext_net1 up &&` + "\n" + `docker exec kind-worker2 ip link set $ifw2 down &&` + "\n" + `docker exec kind-worker2 ip link set $ifw2 name ext_net1 &&` + "\n" + `docker exec kind-worker2 ip link set ext_net1 up)`)
	r.Run(`kubectl create ns nsm-system`)
	r.Run(`cat > patch-nse.yaml <<EOF` + "\n" + `---` + "\n" + `apiVersion: apps/v1` + "\n" + `kind: Deployment` + "\n" + `metadata:` + "\n" + `    name: nse-remote-vlan` + "\n" + `spec:` + "\n" + `  template:` + "\n" + `    spec:` + "\n" + `      containers:` + "\n" + `        - name: nse` + "\n" + `          env:` + "\n" + `          - name: NSM_CONNECT_TO` + "\n" + `            value: "registry:5002"` + "\n" + `          - name: NSM_SERVICES` + "\n" + `            value: "finance-bridge { vlan: 100; via: gw1}"` + "\n" + `          - name: NSM_CIDR_PREFIX` + "\n" + `            value: 172.10.0.0/24,100:200::/64` + "\n" + `          - name: NSM_MAX_TOKEN_LIFETIME` + "\n" + `            value: "60s"` + "\n" + `EOF`)
	r.Run(`kubectl apply -k .`)
	r.Run(`kubectl -n nsm-system wait --for=condition=ready --timeout=2m pod -l app=nse-remote-vlan`)
	r.Run(`WH=$(kubectl get pods -l app=admission-webhook-k8s -n nsm-system --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')` + "\n" + `kubectl wait --for=condition=ready --timeout=1m pod ${WH} -n nsm-system`)
	s.RunIncludedSuites()
}
func (s *Suite) RunIncludedSuites() {
	runTest := func(subSuite suite.TestingSuite, suiteName, testName string, subtest func()) {
		type runner interface {
			Run(name string, f func()) bool
		}
		defer func() {
			if afterTestSuite, ok := subSuite.(suite.AfterTest); ok {
				afterTestSuite.AfterTest(suiteName, testName)
			}
			if tearDownTestSuite, ok := subSuite.(suite.TearDownTestSuite); ok {
				tearDownTestSuite.TearDownTest()
			}
		}()
		if setupTestSuite, ok := subSuite.(suite.SetupTestSuite); ok {
			setupTestSuite.SetupTest()
		}
		if beforeTestSuite, ok := subSuite.(suite.BeforeTest); ok {
			beforeTestSuite.BeforeTest(suiteName, testName)
		}
		// Run test
		subSuite.(runner).Run(testName, subtest)
	}
	s.Run("Rvlanovs", func() {
		s.rvlanovsSuite.SetT(s.T())
		s.rvlanovsSuite.SetupSuite()
		runTest(&s.rvlanovsSuite, "Rvlanovs", "TestKernel2RVlanBreakout", s.rvlanovsSuite.TestKernel2RVlanBreakout)
		runTest(&s.rvlanovsSuite, "Rvlanovs", "TestKernel2RVlanInternal", s.rvlanovsSuite.TestKernel2RVlanInternal)
		runTest(&s.rvlanovsSuite, "Rvlanovs", "TestKernel2RVlanMultiNS", s.rvlanovsSuite.TestKernel2RVlanMultiNS)
	})
	s.Run("Rvlanvpp", func() {
		s.rvlanvppSuite.SetT(s.T())
		s.rvlanvppSuite.SetupSuite()
		runTest(&s.rvlanvppSuite, "Rvlanvpp", "TestKernel2RVlanBreakout", s.rvlanvppSuite.TestKernel2RVlanBreakout)
		runTest(&s.rvlanvppSuite, "Rvlanvpp", "TestKernel2RVlanInternal", s.rvlanvppSuite.TestKernel2RVlanInternal)
		runTest(&s.rvlanvppSuite, "Rvlanvpp", "TestKernel2RVlanMultiNS", s.rvlanvppSuite.TestKernel2RVlanMultiNS)
	})
}
func (s *Suite) Test() {}
