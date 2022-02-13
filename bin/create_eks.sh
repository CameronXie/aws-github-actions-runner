#!/bin/bash

set -ex

instance_type="t3.medium"
private_subnets=$(aws ec2 describe-subnets \
  --filter Name=tag:Name,Values=actions-runner-vpc/VPC/PrivateSubnet? \
  --query 'Subnets[*].[SubnetId,AvailabilityZone]' \
  --region ${CDK_DEFAULT_REGION})

private_subnet1_id=$(echo $private_subnets | jq .[0][0])
private_subnet1_az=$(echo $private_subnets | jq .[0][1])
private_subnet2_id=$(echo $private_subnets | jq .[1][0])
private_subnet2_az=$(echo $private_subnets | jq .[1][1])

build_dir="$(mktemp -d)"
cd "${build_dir}"

cat >cluster.yml <<-CLUSTER
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: actions-runner-cluster
  region: $CDK_DEFAULT_REGION

vpc:
  subnets:
    private:
      $private_subnet1_az: { id: $private_subnet1_id }
      $private_subnet2_az: { id: $private_subnet2_id }

nodeGroups:
  - name: actions-runner-ng-1
    labels:
      role: workers
    instanceType: $instance_type
    minSize: 0
    maxSize: 5
    desiredCapacity: 1
    ssh:
      allow: false
    privateNetworking: true
    iam:
      attachPolicyARNs:
        - arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy
        - arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
        - arn:aws:iam::aws:policy/ElasticLoadBalancingFullAccess
      withAddonPolicies:
        imageBuilder: true
        appMesh: true
        xRay: true
        cloudWatch: true
        externalDNS: true
        autoScaler: true

  - name: actions-runner-ng-2
    labels:
      role: workers
    instanceType: $instance_type
    minSize: 0
    maxSize: 5
    desiredCapacity: 1
    ssh:
      allow: false
    privateNetworking: true
    iam:
      attachPolicyARNs:
        - arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy
        - arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
        - arn:aws:iam::aws:policy/ElasticLoadBalancingFullAccess
      withAddonPolicies:
        imageBuilder: true
        appMesh: true
        xRay: true
        cloudWatch: true
        externalDNS: true
        autoScaler: true

cloudWatch:
  clusterLogging:
    enableTypes: ["*"]

CLUSTER

eksctl create cluster -f cluster.yml
rm -rf "${build_dir}"
