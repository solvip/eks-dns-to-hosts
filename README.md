eks-dns-to-hosts
===

[![Build Status](https://secure.travis-ci.org/solvip/eks-dns-to-hosts.png)](http://travis-ci.org/solvip/eks-dns-to-hosts)

This is a dirty hack circumventing https://github.com/aws/containers-roadmap/issues/221

The use case that this is solving is, for example:
- You've got an AWS account running a NON-PUBLIC EKS cluster
- You've got another AWS account running CI
- Your EKS and CI accounts are VPC peered

Then, you won't be able to look up the EKS cluster endpoints retrieved from the AWS EKS API since the DNS records are private.
