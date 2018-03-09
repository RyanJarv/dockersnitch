package dockersnitch

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/coreos/go-iptables/iptables"
	"github.com/janeczku/go-ipset/ipset"
)

type IPTables struct {
	Chain     string
	NFQueue   uint16
	Blacklist *ipset.IPSet
	Whitelist *ipset.IPSet
}

func (i *IPTables) Setup() {
	cmd := exec.Command("ipset", "create", "-exist", "dockersnitch_blacklistset", "list:set")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	i.Blacklist, err = ipset.New("dockersnitch_blacklist", "hash:ip", ipset.Params{})
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "flush", "dockersnitch_blacklistset")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "add", "dockersnitch_blacklistset", i.Blacklist.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("ipset", "create", "-exist", "dockersnitch_whitelistset", "list:set")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	i.Whitelist, err = ipset.New("dockersnitch_whitelist", "hash:ip", ipset.Params{})
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "flush", "dockersnitch_whitelistset")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("ipset", "add", "dockersnitch_whitelistset", i.Blacklist.Name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	i.Blacklist.Refresh([]string{})
	i.Whitelist.Refresh([]string{})

	ipt, err := iptables.New()
	if err != nil {
		log.Fatal("iptables: could not initialize iptables")
	}
	if err := ipt.NewChain("filter", i.Chain); err != nil {
		log.Printf("iptables: Chain %s already exists, won't continue with setup", i.Chain)
		return
	}

	if err := ipt.Append("filter", i.Chain, "-m", "set", "--match-set", "dockersnitch_whitelist", "dst", "-j", "ACCEPT"); err != nil {
		log.Printf("iptables: Could not set up Accept rule on chain $s", i.Chain)
		log.Fatal(err)
	}
	if err := ipt.Append("filter", i.Chain, "-m", "set", "--match-set", "dockersnitch_blacklist", "dst", "-j", "REJECT"); err != nil {
		log.Printf("iptables: Could not set up Reject rule on chain $s", i.Chain)
		log.Fatal(err)
	}

	if err := ipt.Append("filter", i.Chain, "-j", "LOG"); err != nil {
		log.Printf("iptables: Could not set up Reject rule on chain $s", i.Chain)
		log.Fatal(err)
	}

	if err := ipt.Append("filter", i.Chain, "-j", "NFQUEUE", "--queue-num", fmt.Sprint(i.NFQueue)); err != nil {
		log.Printf("iptables: Could not set up NFQUEUE rule on chain $s", i.Chain)
		log.Fatal(err)
	}

	if err := ipt.Insert("filter", "DOCKER-USER", 1, "-p", "tcp", "-s", "172.17.0.2", "-j", i.Chain); err != nil {
		log.Printf("iptables: Could not set up NFQUEUE rule on chain $s", i.Chain)
		log.Fatal(err)
	}
}

func (i *IPTables) Teardown() {
	ipt, err := iptables.New()
	if err != nil {
		log.Fatalf("iptables: could not initialize iptables")
	}
	if err := ipt.Delete("filter", "DOCKER-USER", "-p", "tcp", "-s", "172.17.0.2", "-j", i.Chain); err != nil {
		log.Printf("iptables: Could not delete rule %s jump rule in DOCKER-USER chain", i.Chain)
	}
	if err := ipt.ClearChain("filter", i.Chain); err != nil {
		log.Printf("iptables: Could not clear chain %s", i.Chain)
		log.Fatal(err)
	}
	if err := ipt.DeleteChain("filter", i.Chain); err != nil {
		log.Printf("iptables: Could not delete chain %s", i.Chain)
		log.Fatal(err)
	}
}
