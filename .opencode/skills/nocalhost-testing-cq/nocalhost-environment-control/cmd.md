# Nocalhost nhctl Commands Reference

## nhctl install
```bash
nhctl install <app-name> \
  -n <namespace> \
  --type rawManifestLocal \
  --local-path <projectPath> \
  --outer-config <appConfig> \
  --kubeconfig <kubeConfig>
```

**Parameters:**
| From Config | Flag | Example |
|-------------|------|---------|
| `developerName` + `origDeployName` | `<app-name>` | `xihe-server-chenqi479` |
| `namespace` | `-n` | `xihe-test-v2` |
| `projectPath` | `--local-path` | `/home/chenqi252/code/prompt-competition/xihe-server-superpowers` |
| `appConfig` | `--outer-config` | `.nocalhost/app.yaml` |
| `kubeConfig` | `--kubeconfig` | `~/.kube/xihe-test-v2_kubeconfig` |

---

## nhctl dev start
```bash
nhctl dev start <app-name> \
  -n <namespace> \
  -d <origDeployName> \
  --dev-mode <duplicate|replace> \
  --image golang:1.24 \
  --kubeconfig <kubeConfig> \
  --without-sync \
  --without-terminal \
  -s <projectPath>
```

**Parameters:**
| From Config | Flag | Example |
|-------------|------|---------|
| `developerName` + `origDeployName` | `<app-name>` | `xihe-server-chenqi479` |
| `namespace` | `-n` | `xihe-test-v2` |
| `origDeployName` | `-d` | `xihe-server` |
| `projectPath` | `-s` | `/home/chenqi252/code/prompt-competition/xihe-server-superpowers` |
| `kubeConfig` | `--kubeconfig` | `~/.kube/xihe-test-v2_kubeconfig` |

---

## nhctl dev end
```bash
nhctl dev end <app-name> \
  -n <namespace> \
  -d <origDeployName> \
  --kubeconfig <kubeConfig>
```

---

## nhctl dev reset
```bash
nhctl dev reset <app-name> \
  -n <namespace> \
  -d <origDeployName> \
  --kubeconfig <kubeConfig>
```

---

## nhctl uninstall
```bash
nhctl uninstall <app-name> \
  -n <namespace> \
  --kubeconfig <kubeConfig>
```

---

## nhctl list
```bash
nhctl list \
  -n <namespace> \
  --kubeconfig <kubeConfig>
```

---

## kubectl patch (post-install security fix)
```bash
kubectl patch deployment <deployName> \
  -n <namespace> \
  --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/securityContext/runAsUser","value":0},{"op":"add","path":"/spec/template/spec/securityContext/capabilities/add","value":["DAC_OVERRIDE"]}]' \
  --kubeconfig <kubeConfig>
```

---

## kubectl rollout restart
```bash
kubectl rollout restart deployment <deployName> \
  -n <namespace> \
  --kubeconfig <kubeConfig>
```

---

## kubectl port-forward
```bash
kubectl port-forward \
  -n <namespace> \
  <podName> \
  <localPort>:<remotePort> \
  --kubeconfig <kubeConfig>
```

---

## kubectl exec (sync files)
```bash
# Pack
tar -czf - --exclude=.git --exclude=vendor <files> | kubectl exec -i -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- tar -xzf - -C /home/nocalhost-dev/

# Or with vendor
tar -czf - --exclude=.git <files> | kubectl exec -i -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- tar -xzf - -C /home/nocalhost-dev/
```

---

## kubectl exec (build)
```bash
kubectl exec -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- bash /home/nocalhost-dev/.nocalhost/build.sh
```

---

## kubectl exec (run)
```bash
kubectl exec -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- bash -c 'export DEVELOPER_NAME=<developerName>; nohup bash /home/nocalhost-dev/.nocalhost/startup.sh > server.log 2>&1 &'
```

---

## kubectl exec (logs)
```bash
kubectl exec -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- tail -f /home/nocalhost-dev/server.log
```

---

## kubectl exec (stop)
```bash
kubectl exec -n <namespace> <podName> -c nocalhost-dev --kubeconfig <kubeConfig> -- pkill <binaryName>
```
