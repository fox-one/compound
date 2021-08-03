# Deployment

All the deploying instructions are in the [Makefile](./Makefile)

#### Environments：

* local  
* test  
* prod

Create the config file named as `config.${ENV}.yaml` according to the [template file](./deploy/../../deploy/config.node.yaml.tpl) and place it in the deploy directory`./deploy/`, such as：`config.local.yaml, config.test.yaml, config.prod.yaml`。

## Build as executable file locally

* build

```
% make build
```

* run

```
% ./builds/compound server --port 8010 --config ./config/config.yaml
% ./builds/compound worker --port 8020 --config ./config/config.yaml
```

## Build as docker image locally
* Place `config.audit-node1.yaml`,`config.audit-node2.yaml`,`config.audit-node3.yaml` in the dir `./deploy`
* Build the docker images

```
% make REPOSITORY_PATH='' docker-build-audit-node1 
% make REPOSITORY_PATH='' docker-build-audit-node2
% make REPOSITORY_PATH='' docker-build-audit-node3
```
* Config the [docker compse file](./deploy/docker/docker-compose.yml)
  
* Run docker compose

```
% docker-compse -f ./deploy/docker/docker-compse.yml up -d
```

## Build as docker image with `REPOSITORY_PATH` 
* Build the docker image(put the config into the docker image)：
  1. Modify the `REPOSITORY_PATH` in Makefile
  2. deploy the image to the docker repository, execute `make deploy-%`, E.g: `make deploy-prod`


* If you want to load the config outside the docker image:
  1. Modify `Dockerfile`, delete `ADD config/config.yaml config/config.yaml`
  2. Modify `Dockerfile`，add  `VOLUME [ "/var/data/compound" ]`
  3. Put the config file to the host directory `/var/data/compound`

* health check api
   1. api:   `/hc`
   2. worker: `/hc`

## Run

* Run api server

```

// port: custom the api port，default is 80，
// config: custom the config file， default path is ./config/config.yaml
./builds/compound server --port 80 --config ./config/config.yaml

```

* Run worker server

```
// config: custom the config file， default path is ./config/config.yaml
./builds/compound worker --config ./config/config.yaml
```

* Preparing before business running
  
> This project is just a node layer that only including the core business and core apis. And the application layer is `https://github.com/fox-one/compound-app`, which contains the user interaction apis and view logics.

> All management-related instructions reference documents [governance](./docs/governance.md)

  1. Deposit some vote token to the multi-sign wallet with `./builds/compound deposit --asset xxx --amount`
     ```
      vote:
         asset: 965e5c6e-434c-3fa9-b780-c50f43cd955c
         amount: 0.00000001
     ```
  2. Deposit ctoken to the miltisign wallet with `./builds/compound deposit --asset xxx --amount`
  3. Init market data with `./builds/compound market .....`
  4. Add price oracle signer that provide the market price with `./builds/compound add-oracle-signer`