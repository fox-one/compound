# compound

> compound 节点引擎。

## 项目结构说明

* **./Makefile**  项目构建
* **./Dockerfile** 
* **./main.go** 程序主入口
* **./cmd** 
* **./core** 业务模型
* **./store** 数据持久化
* **./service** 业务逻辑
* **./worker** 后台服务
* **./config** 缺省的配置文件目录，[配置模板](./deploy/config.node.yaml.tpl)
* **./deploy** 部署文件目录
    1. **test** 测试环境配置
    2. **prod** 线上环境配置
    3. **local** 本地环境配置


## 部署

### 1. 普通部署

### 2. docker 容器部署

## 服务启动

* 往节点机器人转一定数量的`vote_asset`用于价格投票用及其他投票用