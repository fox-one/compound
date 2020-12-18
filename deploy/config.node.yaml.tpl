genesis: 1603382400
location: Asia/Shanghai

db:
  dialect: mysql
  host: ~
  read_host: ~
  port: 3306
  user: ~
  password: ~
  database: ~
  location: Asia%2FShanghai
  Debug: true

# 每个节点部署一个自己的price oracle
price_oracle:
  end_point: https://poracle-dev.fox.one

dapp:
  num: 7000103159
  client_id: ~
  session_id: ~
  client_secret: ~
  pin_token: ~
  pin: ""
  private_key: ~

group:
# 所有节点共享的私钥
  private_key: ~
  # 节点用户签名的私钥
  sign_key: ~
  # 该节点的管理员
  admins:
    - ~
    - ~
    - ~ 
  # 节点成员 
  members:
    - client_id: ~
    # 节点用于校验签名的公钥
      verify_key: ~
  threshold: 2
  vote:
    asset: 965e5c6e-434c-3fa9-b780-c50f43cd955c
    amount: 0.00000001