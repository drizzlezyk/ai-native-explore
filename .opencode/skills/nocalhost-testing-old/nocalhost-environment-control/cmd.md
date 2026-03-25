# Install nocalhost app (first time only)
```bash
nhctl install xihe-server-$XIHE_USERNAME -n xihe-test-v2 \
    --type rawManifestLocal \
    --local-path $PROJECT_DIR \
    --outer-config .nocalhost/app.yaml \
    --kubeconfig $KUBECONFIG
```


nhctl install xihe-server-$XIHE_USERNAME -n xihe-test-v2 \
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
    - name: xihe-server  
      serviceType: deployment  
      containers:  
        - name: xihe-server  
          dev:  
            image: golang:1.24  
            shell: /bin/bash
```


```yaml
# $PROJECT_DIR/.nocalhost/config.yaml 
name: xihe-server
resourcePath: ["deployments/xihe-server"]
serviceType: deployment
containers:
  - name: xihe-server
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
nhctl uninstall xihe-server-$XIHE_USERNAME  -n xihe-test-v2 --kubeconfig $KUBECONFIG


nhctl list xihe-server-$XIHE_USERNAME -n xihe-test-v2 --kubeconfig $KUBECONFIG



nhctl dev end xihe-server-${XIHE_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG


nhctl dev reset xihe-server-${XIHE_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG
```


XIHE_USERNAME=chenqi479

KUBECONFIG=~/

nhctl dev start xihe-server-$XIHE_USERNAME -n xihe-test-v2 \
    --dev-mode duplicate \
    -s . \
    -d xihe-server \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  


# Start dev mode in duplicate mode (requires -s for local sync path and -d for deployment)
```bash
output=$(nhctl dev start xihe-server-$XIHE_USERNAME -n xihe-test-v2 \
    --dev-mode duplicate \
    -s . \
    -d xihe-server \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  2>&1)

echo "$output"
```

```
Starting duplicate DevMode...
[name: xihe-server serviceType: deployment]                            Success load svc config from local file [/home/chenqi252/code/prompt-competition/xihe-server-superpowers/.nocalhost/config.yaml]
Disabling hpa...
Failed to find hpa: : horizontalpodautoscalers.autoscaling is forbidden: User "system:serviceaccount:xihe-test-v2:chenqi-developer-sa" cannot list resource "horizontalpodautoscalers" in API group "autoscaling" in the namespace "xihe-test-v2"
No hpa found
Mount workDir to emptyDir
[WARNING] Resources Limits: 1 cpu, 1000Mi memory is less than the recommended minimum: 2 cpu, 2Gi memory. Running programs in DevContainer may fail. You can increase Resource Limits in Nocalhost Config
Creating xihe-server-i24-1-586e7910(apps/v1, Kind=Deployment)
Resource(Deployment) xihe-server-i24-1-586e7910 created
Patching [{"op":"replace","path":"/spec/replicas","value":1}]
deployment.apps/xihe-server-i24-1-586e7910 patched (no change)
Now waiting dev mode to start...

Pod xihe-server-i24-1-586e7910-75ff455bf-4s4vl now Pending
 * Condition: ContainersNotInitialized, containers with incomplete status: [vault-agent-init]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod xihe-server-i24-1-586e7910-75ff455bf-4s4vl now Pending
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod xihe-server-i24-1-586e7910-75ff455bf-4s4vl now Running
 >> Container: nocalhost-dev is Running
 >> Container: nocalhost-sidecar is Running

deployment.apps/xihe-server patched
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
nhctl dev end xihe-server-${XIHE_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG

nhctl uninstall xihe-server-$XIHE_USERNAME  -n xihe-test-v2 --kubeconfig $KUBECONFIG
```