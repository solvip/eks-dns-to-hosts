eks-dns-to-hosts
===

[![Build Status](https://secure.travis-ci.org/solvip/eks-dns-to-hosts.png)](http://travis-ci.org/solvip/eks-dns-to-hosts)

This is a dirty hack circumventing https://github.com/aws/containers-roadmap/issues/221

eks-dns-to-hosts will update `/etc/hosts`, maintaining entries such as:
```
172.23.2.35	ENDPOINT-ID.yl4.eu-west-1.eks.amazonaws.com
172.23.5.88	ENDPOINT-ID.yl4.eu-west-1.eks.amazonaws.com
```

If the ip addresses change, such as on cluster update, they will be replaced on the next `eks-dns-to-hosts` run for that cluster.

# Why?

One use case is, for example:
- You've got an AWS account running a NON-PUBLIC EKS cluster
- You've got another AWS account running CI
- Your EKS and CI accounts are VPC peered

Then, you won't be able to look up the EKS cluster endpoints retrieved from the AWS EKS API since the DNS records are private.

# Installation

We've got pre-built releases for Linux.  You can do a:

```sh
# wget -O /usr/local/bin/eks-dns-to-hosts https://github.com/solvip/eks-dns-to-hosts/releases/download/v1.0.0/eks-dns-to-hosts
# chown root /usr/local/bin/eks-dns-to-hosts
# chmod +s /usr/local/bin/eks-dns-to-hosts
```

# Usage

```sh
$ eks-dns-to-hosts my-eks-cluster
```
