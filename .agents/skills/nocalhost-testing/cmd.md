# Install nocalhost app (first time only)
```bash
nhctl install xihe-server-$GITHUB_USERNAME -n xihe-test-v2 \
    --type rawManifestLocal \
    --local-path $PROJECT_DIR \
    --outer-config .nocalhost/app.yaml \
    --kubeconfig $KUBECONFIG
```

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
nhctl uninstall xihe-server-$GITHUB_USERNAME  -n xihe-test-v2 --kubeconfig $KUBECONFIG


nhctl list xihe-server-$GITHUB_USERNAME -n xihe-test-v2 --kubeconfig $KUBECONFIG



nhctl dev end xihe-server-${GITHUB_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG


nhctl dev reset xihe-server-${GITHUB_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG
```

# Start dev mode in duplicate mode (requires -s for local sync path and -d for deployment)
```bash
output=$(nhctl dev start xihe-server-$GITHUB_USERNAME -n xihe-test-v2 \
    --dev-mode duplicate \
    -s $PROJECT_DIR \
    -d xihe-server \
    --image golang:1.24 \
    --kubeconfig $KUBECONFIG \
    --without-terminal \
    --without-sync  2>&1)

echo "$output"
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
nhctl dev end xihe-server-${GITHUB_USERNAME} -n xihe-test-v2 -d $DEPLOY_NAME --kubeconfig $KUBECONFIG

kubectl delete deployment $DEPLOY_NAME -n xihe-test-v2 --kubeconfig $KUBECONFIG

nhctl uninstall xihe-server-$GITHUB_USERNAME  -n xihe-test-v2 --kubeconfig $KUBECONFIG
```