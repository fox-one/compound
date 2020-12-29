# compound

> compound 节点引擎。

## 项目结构

```

---
|-cmd     // cli入口 
|-config  //缺省的配置文件目录
|-deploy  //部署目录
|-docs    //存放项目文档
|-core  //业务数据模型
|-pkg     //公共库目录
|-service //业务
|-store   //数据持久化层
|-worker  //后台服务
|-handler     // api服务
|-Dockerfile // docker 配置文件
|-main.go //程序入口

```

## 配置
 参考[配置模板](deploy/config.node.yaml.tpl)

## 运行

* 运行API服务

```

// port: 自定义端口，缺省端口为80，
// config: 自定义配置文件， 缺省路径为 ./config/config.yaml
./compound server --port 80 --config ./config/config.yaml

```

* 运行worker

```
// config: 自定义配置文件，缺省路径为 ./config/config.yaml
./compound worker --config ./config/config.yaml
```

> 注意：运行worker 前需要给节点DAPP转一些`Vote asset`过去，因为每个节点都会消费`Vote asset`往链上写价格


## 部署

部署文件查看[Makefile](./Makefile)

部署环境有3个：
* local  //本地环境
* test   //测试环境
* prod   //生产环境

对应的在deploy目录下有3个环境对应的配置文件`config.${ENV}.yaml`，如下：`config.local.yaml, config.test.yaml, config.prod.yaml`。


* 直接编译可执行文件到本地：
```
make build-%    //如：make build-local
```

* 编译为docker镜像(配置文件打包进docker镜像)：
  1. 修改 Makefile `REPOSITORY_PATH`的值
  2. 发布镜像到仓库执行 `make deploy-%`，例：make deploy-prod


* 如不想把配置文件打包进docke镜像
  1. 修改`Dockerfile`, 删除 `ADD config/config.yaml config/config.yaml`
  2. 修改`Dockerfile`，增加配置 `VOLUME [ "/var/data/compound" ]`
  3. 把配置文件放在host目录下 `/var/data/compound`
  4. 运行时通过config配置自定义配置文件，如：`./compound server --port 80 --config /var/data/compound/config.yaml`

* health check 接口
   1. api:   `/hc`
   2. worker: `/hc` 