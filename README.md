# cd-tools
创建或升级集群内的Deployment，如果集群内已经存在则进行升级，同时保留原始yaml文件，如不存在，则进行创建。 失败后会回滚到之前版本。 注意目前仅支持Deployment类型。
