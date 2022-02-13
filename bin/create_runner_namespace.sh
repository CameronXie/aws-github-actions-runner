#!/bin/bash

set -ex

launcherRole=$(aws lambda get-function --function-name actions-runner-launcher --query Configuration.Role --region ${CDK_DEFAULT_REGION})
terminatorRole=$(aws lambda get-function --function-name actions-runner-terminator --query Configuration.Role --region ${CDK_DEFAULT_REGION})
token=$(echo -n ${GITHUB_TOKEN} | base64 -w 0)

build_dir="$(mktemp -d)"
cd "${build_dir}"

cat >runner_ns.yml <<-RUNNER_NS
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-auth
  namespace: kube-system
data:
  mapUsers: |
    - userarn: ${launcherRole}
      username: launcher
    - userarn: ${terminatorRole}
      username: terminator
---
apiVersion: v1
kind: Namespace
metadata:
  name: actions-runner
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: actions-runner
  name: launcher
rules:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "watch", "list", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: actions-runner
  name: launcher-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: launcher
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: launcher
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: actions-runner
  name: terminator
rules:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "watch", "list", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: actions-runner
  name: terminator-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: terminator
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: terminator
---
apiVersion: v1
kind: Secret
metadata:
  namespace: actions-runner
  name: github-token
type: Opaque
data:
  GITHUB_PERSONAL_TOKEN: ${token}

RUNNER_NS

kubectl apply -f runner_ns.yml
