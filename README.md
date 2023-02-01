# cd-tools
创建或升级集群内的Deployment，如果集群内已经存在则进行升级，同时保留原始yaml文件，如不存在，则进行创建。 失败后会回滚到之前版本。 注意目前仅支持Deployment类型。

### 使用方法

./cd-tools apply xxx

#### 参数说明

```shell
Usage:
  cd-tools apply [flags]

Flags:
  -n, --cluster-name string   设置集群名称，默认为master ip
  -f, --file string           指定创建或更新的K8S yaml文件路径
  -h, --help                  help for apply
  -c, --kube-config string    指定KubeConfig文件 (default "/root/.kube/config")
  -t, --timeout string        指定超时时间，单位"ns"、“us”、“µs”、“ms”、“s”、“m”、“h” (default "60s")
  -w, --workdir string        工作路径，用来存放备份文件 (default "/root/.cd-tools")

```

#### 示例
```shell
./cd-tools apply -c /Users/liuchen/.kube/config -f testdata/1.yml -w testdata/workdir -t 60s
```