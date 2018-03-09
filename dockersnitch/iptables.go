package dockersnitch

import (
	"fmt"
	"log"

	"github.com/coreos/go-iptables/iptables"
)

const IPTChain = "DOCKERSNITCH"

func SetupIPTables() {
	ipt, err := iptables.New()
	if err != nil {
		log.Fatal("iptables: could not initialize iptables")
	}
	if err := ipt.NewChain("filter", IPTChain); err != nil {
		log.Printf("iptables: Chain %s already exists, won't continue with setup", IPTChain)
		return
	}

	if err := ipt.Append("filter", IPTChain, "-m", "set", "--match-set", "dockersnitch_whitelist", "dst", "-j", "ACCEPT"); err != nil {
		log.Printf("iptables: Could not set up Accept rule on chain $s", IPTChain)
		log.Fatal(err)
	}
	if err := ipt.Append("filter", IPTChain, "-m", "set", "--match-set", "dockersnitch_blacklist", "dst", "-j", "REJECT"); err != nil {
		log.Printf("iptables: Could not set up Reject rule on chain $s", IPTChain)
		log.Fatal(err)
	}

	//if err := ipt.Append("filter", IPTChain, "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT"); err != nil {
	//	log.Printf("iptables: Could not set up ACCEPT rule on chain $s", IPTChain)
	//	log.Fatal(err)
	//}

	if err := ipt.Append("filter", IPTChain, "-j", "LOG"); err != nil {
		log.Printf("iptables: Could not set up Reject rule on chain $s", IPTChain)
		log.Fatal(err)
	}

	if err := ipt.Append("filter", IPTChain, "-j", "NFQUEUE", "--queue-num", fmt.Sprint(NFQueueNum)); err != nil {
		log.Printf("iptables: Could not set up NFQUEUE rule on chain $s", IPTChain)
		log.Fatal(err)
	}

	if err := ipt.Insert("filter", "DOCKER-USER", 1, "-p", "tcp", "-s", "172.17.0.2", "-j", IPTChain); err != nil {
		log.Printf("iptables: Could not set up NFQUEUE rule on chain $s", IPTChain)
		log.Fatal(err)
	}
}

func TeardownIPTables() {
	ipt, err := iptables.New()
	if err != nil {
		log.Fatalf("iptables: could not initialize iptables")
	}
	if err := ipt.Delete("filter", "DOCKER-USER", "-p", "tcp", "-s", "172.17.0.2", "-j", IPTChain); err != nil {
		log.Printf("iptables: Could not delete rule %s jump rule in DOCKER-USER chain", IPTChain)
	}
	if err := ipt.ClearChain("filter", IPTChain); err != nil {
		log.Printf("iptables: Could not clear chain %s", IPTChain)
		log.Fatal(err)
	}
	if err := ipt.DeleteChain("filter", IPTChain); err != nil {
		log.Printf("iptables: Could not delete chain %s", IPTChain)
		log.Fatal(err)
	}
}
