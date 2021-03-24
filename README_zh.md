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


## 管理员工具

### Keys
> 生成Ed25519公私钥对

cmd:

```
./compound keys
```

### inject-ctoken
> 往多签钱包注入ctoken

cmd:

```
./compound inject-ctoken --asset xxxxx --amount 10000
or
./compound ic --asset xxxxx --amount 10000
```

### withdraw
> 发起从多签钱包提现的投票请求

cmd:

```
./compound withdraw --opponent xxxx --asset xxxxxxx --amount 10000
```

### add-market
> 添加借贷市场

cmd:

```
//-s  symbol
//-a asset_id
//-c ctoken_asset_id
./compound add-market --s BTC --a xxxxxx --c yyyyyy
or
./compound am -s BTC -a xxxxx -c yyyyyyy
```

### update-market
> 更新market参数

cmd:

```
//-s symbol
//-ie init_exchange
//-rf reserve_factor
//-li liquidation_incentive
//-cf collateral_factor
//-br base_rate
./compound update-market --s BTC --ie 1 --rf 0.1 --li 0.05 --cf 0.75 --br 0.025
or
./compound um --s BTC --ie 1 --rf 0.1 --li 0.05 --cf 0.75 --br 0.025
```

### update-market-advance
> 更新market参数

cmd:

```
//-s symbol
//-bc borrow_cap
//-clf close_factor
//-m multiplier
//-jm jump_multiplier
//-k kink
./compound update-market-advance --s BTC --bc 0 --clf 0.5 --m 0.3 --jm 0.5 --k 0.7
or
./compound uma --s BTC --bc 0 --clf 0.5 --m 0.3 --jm 0.5 --k 0.7
```

### close-market
> 关闭market

cmd:

```
./compound close-market --asset xxxxxxxx
or
./compound cm --asset xxxxxx
```

### open-market
> open market

cmd:

```
./compound open-market --asset xxxxx
or
./compound om --asset xxxxxxx
```

## 附

### Price oracle(价格预言机)

> compound 的price oracle通过以下机制来保证compound market价格的稳定：

* 每个compound节点都像链上提供价格
* compound通过m/n多签的方式，保证n个节点至少m个节点提供的价格有效才算该market的当前block的价格有效，算法如下：
  1. 把所有节点提供的价格按价格升序排序
  2. 相邻价格比较，差值>=5%的价格无效
  3. 剩余价格满足m/n则此次价格有效
  4. 把m个价格算均值a，a就是market当前blockNum的价格
* 每个节点是通过price oracle服务获取指定market 1小时内的价格均值作为当前节点提供的价格写上主链
* 当出现价格被恶意攻击时，由于是用1小时内的价格均值作为market的价格，所以恶意的价格对market的影响有限，同时管理人员可以同时决策是否有必要`close-market`，并在价格恢复后`open-market`

### market 熔断
> 某个market价格异常情况下，关闭市场

* 当某个market的价格被恶意攻击时，管理人员有权执行`close-market`命令，申请闭市投票，投票通过则market关闭
* 已关闭的market不可借贷，其他market不影响正常借贷
* 但是当存在至少一个market关闭的情况下，将禁止所有market的清算行为，因为清算会影响到用户所有账户的流动性
