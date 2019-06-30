package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Printf("unable to create aws session: %s\n", err)
		os.Exit(1)
	}

	eksSvcClient := eks.New(sess)
	ec2SvcClient := ec2.New(sess)

	endpoint, err := getClusterEndpoint(eksSvcClient, "kubedev")
	if err != nil {
		fatalf("unable to get cluster endpoint: %s\n", err)
	} else if endpoint == "" {
		fatalf("empty cluster endpoint returned from aws\n")
	}

	ips, err := getClusterControlPlaneIPs(ec2SvcClient, "kubedev")
	if err != nil {
		fatalf("unable to get cluster control plane ip addresses: %s\n", err)
	}

	hosts, err := readHostsIntoLines()
	if err != nil {
		fatalf("unable to read /etc/hosts: %s\n", err)
	}

	// Create a temporary file which we'll write into; we'll then rename it to /etc/hosts
	newHostsFilename, err := writeNewHostsToTempFile(hosts, endpoint, ips)
	if err != nil {
		fatalf("unable to write new, temporary, hosts file: %s\n", err)
	}
	defer os.Remove(newHostsFilename)
	if err := os.Chmod(newHostsFilename, 0644); err != nil {
		fatalf("unable to set permissions of new hosts file to 0644: %s\n", err)
	}

	if err := os.Rename(newHostsFilename, "/etc/hosts"); err != nil {
		fatalf("unable to update /etc/hosts: %s\n", err)
	}
}

// writeNewHostsToTempFile - Write a new hosts file into a temporary file
// If we find an occurance of 'endpoint' in the current hosts file, then we skip those lines before adding
// the new ones
func writeNewHostsToTempFile(curHosts []string, endpoint string, ips []string) (filename string, err error) {
	f, err := ioutil.TempFile("/tmp/", "eks-dns-to-hosts.")
	if err != nil {
		fatalf("unable to create temporary file: %s\n", err)
	}
	defer f.Close()

	bufW := bufio.NewWriter(f)

	// Write the current hosts; skipping previous endpoints
	for _, line := range curHosts {
		if strings.Contains(line, endpoint) {
			continue
		}
		fmt.Fprintln(bufW, line)
	}

	// Write the new endpoints
	for _, ip := range ips {
		fmt.Fprintf(bufW, "%s\t%s\n", ip, endpoint)
	}

	if err := bufW.Flush(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("unable to flush output to temporary file %s: %s", f.Name(), err)
	}

	return f.Name(), nil
}

func readHostsIntoLines() (lines []string, err error) {
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lineScanner := bufio.NewScanner(file)
	for lineScanner.Scan() {
		lines = append(lines, lineScanner.Text())
	}
	if lineScanner.Err() != nil {
		return nil, fmt.Errorf("unable to scan lines: %s", err)
	}

	return lines, nil
}

func fatalf(fmtString string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtString, args...)
	os.Exit(1)
}

func getClusterEndpoint(eksSvcClient *eks.EKS, name string) (endpoint string, err error) {
	out, err := eksSvcClient.DescribeCluster(&eks.DescribeClusterInput{Name: &name})
	if err != nil {
		return "", fmt.Errorf("unable to call eks.DescribeCluster: %s", err)
	}

	if out.Cluster == nil {
		return "", fmt.Errorf("eks.DescribeCluster returned no cluster with name = '%s'", name)
	}

	if out.Cluster.Endpoint == nil {
		return "", fmt.Errorf("eks.DescribeCluster returned nil Endpoint")
	}
	rawEndpoint := *out.Cluster.Endpoint

	parsedURL, err := url.Parse(rawEndpoint)
	if err != nil {
		return "", fmt.Errorf("unable to URL parse endpoint = '%s': %s", rawEndpoint, err)
	}

	return parsedURL.Host, nil
}

// getClusterControlPlaneIPs - return a slice of ip addresses
func getClusterControlPlaneIPs(ec2SvcClient *ec2.EC2, eksClusterName string) (ips []string, err error) {
	networkInterfaceName := fmt.Sprintf("Amazon EKS %s", eksClusterName)
	filter := new(ec2.Filter).SetName("description").SetValues([]*string{&networkInterfaceName})
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{filter},
	}

	out, err := ec2SvcClient.DescribeNetworkInterfaces(input)
	if err != nil {
		return nil, fmt.Errorf("unable to call ec2.DescribeNetworkInterfaces: %s", err)
	}

	for _, v := range out.NetworkInterfaces {
		if v.PrivateIpAddress != nil {
			ips = append(ips, *v.PrivateIpAddress)
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no interfaces with description = '%s' found", networkInterfaceName)
	}

	return ips, nil
}
