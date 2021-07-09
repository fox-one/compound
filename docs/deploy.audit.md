Just for audit example nodes

### Steps:

* Go to the project root directory
* Place the configuration files(`config.audit-node1.yaml`, `config.audit-node2.yaml`, `config.audit-node3.yaml`) in the deployment directory `./deploy`
* Build and generate docker image `make docker-build`
* Go to the docker deployment directory `./deploy/docker`,  `cd deploy/docker`
* Run `docker-compose up db`
* Then connect to the db with root user and grant privileges to `compound` user:
  ```
  mysql> grant all privileges on *.* to compound@'%' identified by 'compound';
  mysql> flush privileges;
  ```
* Then connect to the db with `compound` user and create db `compound_node1`,`compound_node2`,`compound_node3`
* And then you can start up compound services:
  ```
  docker-compose up compound-node1-worker
  docker-compose up compound-node1-server

  docker-compose up compound-node2-worker
  docker-compose up compound-node2-server

  docker-compose up compound-node3-worker
  docker-compose up compound-node3-worker
  
  ```