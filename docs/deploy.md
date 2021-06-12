# Deployment

## Run

* Run api server

```

// port: custom the api port，default is 80，
// config: custom the config file， default path is ./config/config.yaml
./compound server --port 80 --config ./config/config.yaml

```

* Run worker server

```
// config: custom the config file， default path is ./config/config.yaml
./compound worker --config ./config/config.yaml
```

> Notice：Before run worker server should transfer some `Vote asset` to node dapp bot for providing price to the chain.


## Deployment

[Makefile](./Makefile)

Environments：

* local  
* test  
* prod

Put the environment config file named as `config.${ENV}.yaml` in the deploy directory, such as：`config.local.yaml, config.test.yaml, config.prod.yaml`。


* Build the program locally：
```
make build    //E.g: make build-local
```

* Build docker image(put the config into the docker image)：
  1. Modify the `REPOSITORY_PATH` in Makefile
  2. deploy the image to the docker repository, execute `make deploy-%`, E.g: `make deploy-prod`


* If you want to load the config outside the docker image:
  1. Modify `Dockerfile`, delete `ADD config/config.yaml config/config.yaml`
  2. Modify `Dockerfile`，add  `VOLUME [ "/var/data/compound" ]`
  3. Put the config file to the host directory `/var/data/compound`
  4. Run program，E.g: `./compound server --port 80 --config /var/data/compound/config.yaml`

* health check api
   1. api:   `/hc`
   2. worker: `/hc`