# Install nocalhost app (first time only)

{
  "appName": "chenqi47",
  "kubeConfig": "/home/chenqi252/.kube/infra-hk-test-cluster-001-openeuler-bigfiles-kubeconfig",
  "namespace": "openeuler-bigfiles",
  "appConfig": ".nocalhost/app.yaml",
  "deployConfig": ".nocalhost/config.yaml",
  "startupScript": ".nocalhost/startup.sh",
  "buildScript": ".nocalhost/build.sh",
  "heartbeatUrl": "http://localhost:5000/",
  "origDeployName": "openeuler-bigfiles-deployment",
  "binaryName": "main"
}

export KUBECONFIG=/home/chenqi252/.kube/infra-hk-test-cluster-001-openeuler-bigfiles-kubeconfig

export namespace=openeuler-bigfiles

export odn=openeuler-bigfiles-deployment
export app_name=openeuler-bigfiles-deployment-chenqi47

```bash
nhctl install openeuler-bigfiles-deployment-chenqi47 -n ${namespace} \
    --type rawManifestLocal \
    --local-path /home/chenqi252/code/nocalhost-skill/BigFiles \
    --outer-config .nocalhost/app.yaml \
    --kubeconfig $KUBECONFIG
```


nhctl install ${app_name} -n ${namespace}\
    --type rawManifestLocal \
    --local-path . \
    --outer-config .nocalhost/app.yaml \
    --kubeconfig $KUBECONFIG

```yaml
# $PROJECT_DIR/.nocalhost/app.yaml 
configProperties:  
  version: v2  
application:  
  services:  
    - name: ${DEPLOYMENT_NAME}  
      serviceType: deployment  
      containers:  
        - name: ${DEPLOYMENT_NAME}  
          dev:  
            image: golang:1.24  
            shell: /bin/bash
```


```yaml
# $PROJECT_DIR/.nocalhost/config.yaml 
name: ${DEPLOYMENT_NAME}
resourcePath: ["deployments/${DEPLOYMENT_NAME}"]
serviceType: deployment
containers:
  - name: ${DEPLOYMENT_NAME}
    dev:
      gitUrl: ""
      image: golang:1.24
      shell: bash
      workDir: /home/nocalhost-dev
      storageClass: ""
      resources: null
      persistentVolumeDirs: []
      command: null
      debug: null
      sync: null
      env:
        - name: GOCACHE
          value: /home/nocalhost-dev/.cache/go-build
        - name: GOPROXY
          value: https://goproxy.cn,direct
        - name: HOME
          value: /home/nocalhost-dev
      portForward: []
```

```bash
nhctl uninstall ${app_name}  -n ${namespace} --kubeconfig $KUBECONFIG


nhctl list ${app_name} -n ${namespace} --kubeconfig $KUBECONFIG



nhctl dev end ${app_name} -n ${namespace} -d $dn --kubeconfig $KUBECONFIG


nhctl dev reset ${app_name} -n ${namespace} -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n ${namespace} --kubeconfig $KUBECONFIG
```


DEVELOPER_NAME=chenqi479

KUBECONFIG=~/

nhctl dev start ${app_name} -n ${namespace}\
    --dev-mode duplicate \
    -s . \
    -d ${DEPLOYMENT_NAME} \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  


# Start dev mode in duplicate mode (requires -s for local sync path and -d for deployment)
```bash
output=$(nhctl dev start ${app_name} -n ${namespace}\
    --dev-mode duplicate \
    -s /home/chenqi252/code/nocalhost-skill/BigFiles \
    -d $odn \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  2>&1)



kubectl patch deployment $dn -n $namespace --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/securityContext/fsGroup", "value": 1000}]'

kubectl patch deployment $dn -n $namespace --type='json'   -p='[{"op": "add", "path": "/spec/template/spec/securityContext/fsGroup", "value": 1000}]'


echo "$output"
```

```
Starting duplicate DevMode...
[name: ${DEPLOYMENT_NAME} serviceType: deployment]                            Success load svc config from local file [/home/chenqi252/code/prompt-competition/${DEPLOYMENT_NAME}-superpowers/.nocalhost/config.yaml]
Disabling hpa...
Failed to find hpa: : horizontalpodautoscalers.autoscaling is forbidden: User "system:serviceaccount:${NAMESPACE}:chenqi-developer-sa" cannot list resource "horizontalpodautoscalers" in API group "autoscaling" in the namespace "${NAMESPACE}"
No hpa found
Mount workDir to emptyDir
[WARNING] Resources Limits: 1 cpu, 1000Mi memory is less than the recommended minimum: 2 cpu, 2Gi memory. Running programs in DevContainer may fail. You can increase Resource Limits in Nocalhost Config
Creating ${DEPLOYMENT_NAME}-i24-1-586e7910(apps/v1, Kind=Deployment)
Resource(Deployment) ${DEPLOYMENT_NAME}-i24-1-586e7910 created
Patching [{"op":"replace","path":"/spec/replicas","value":1}]
deployment.apps/${DEPLOYMENT_NAME}-i24-1-586e7910 patched (no change)
Now waiting dev mode to start...

Pod ${DEPLOYMENT_NAME}-i24-1-586e7910-75ff455bf-4s4vl now Pending
 * Condition: ContainersNotInitialized, containers with incomplete status: [vault-agent-init]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod ${DEPLOYMENT_NAME}-i24-1-586e7910-75ff455bf-4s4vl now Pending
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod ${DEPLOYMENT_NAME}-i24-1-586e7910-75ff455bf-4s4vl now Running
 >> Container: nocalhost-dev is Running
 >> Container: nocalhost-sidecar is Running

deployment.apps/${DEPLOYMENT_NAME} patched
 ✓  Dev container has been updated
 ✓  File sync is not started caused by --without-sync flag..
```


# Extract deployment name (更精确地提取括号前的名称)
```bash
export DEPLOY_NAME=$(echo "$output" | grep "Kind=Deployment" | head -1 | awk '{print $2}' | cut -d'(' -f1)
```

# Extract pod name (只匹配 Running 状态，确保 Pod 已就绪)
```bash
export POD_NAME=$(echo "$output" | grep "Pod .* now (Running|Pending)" | head -1 | awk '{print $2}')
```


# close
```bash
nhctl dev end ${app_name} -n ${namespace}-d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n ${namespace}--kubeconfig $KUBECONFIG

nhctl uninstall ${app_name}  -n ${namespace}--kubeconfig $KUBECONFIG
```