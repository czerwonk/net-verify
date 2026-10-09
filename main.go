package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"net"
	"os"
)

// NetDefinition defines expectataions for network configuration
type NetDefinition struct {
	Interfaces []*IfaceDefinition `json:"interfaces"`
}

// IfaceDefinition defines expectations to a network interface (e.g. IP address)
type IfaceDefinition struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
}

const version = "0.1.1"

func main() {
	filePath := flag.String("file", "net-definition.json", "JSON file containing the expected network definition")
	showVersion := flag.Bool("version", false, "Prints version info")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := run(*filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("net-verify")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func run(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("could not open file. %v", err)
	}
	defer f.Close()

	d, err := loadDefinition(f)
	if err != nil {
		return err
	}

	return verifyDefinition(d)
}

func loadDefinition(reader io.Reader) (*NetDefinition, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("could not read definition. %v", err)
	}

	d := &NetDefinition{}
	err = json.Unmarshal(b, d)
	if err != nil {
		return nil, fmt.Errorf("could not parse definition. %v", err)
	}

	return d, nil
}

func verifyDefinition(d *NetDefinition) error {
	for _, iface := range d.Interfaces {
		err := verifyInterface(iface)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyInterface(iface *IfaceDefinition) error {
	i, err := net.InterfaceByName(iface.Name)
	if err != nil {
		return fmt.Errorf("%v (%s)", iface.Name, err)
	}

	addrs, err := i.Addrs()
	if err != nil {
		return fmt.Errorf("could not get addresses for interface %s. %v", iface.Name, err)
	}

	for _, expected := range iface.Addresses {
		if !hasIP(expected, addrs) {
			return fmt.Errorf("expected address %s is not assigned to interface %s", expected, iface.Name)
		}
	}

	return nil
}

func hasIP(expected string, addrs []net.Addr) bool {
	for i := range addrs {
		if addrs[i].String() == expected {
			return true
		}
	}

	return false
}
