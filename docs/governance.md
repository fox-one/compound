# Governence

## Tools commands
### Keys
> generate Ed25519 pair

cmd:

```
./compound keys
```

### inject-ctoken
> inject the ctoken to the multi-sign wallet

cmd:

```
./compound inject-ctoken --asset xxxxx --amount 10000
or
./compound ic --asset xxxxx --amount 10000
```

### withdraw
> Initiate a withdrawal proposal

cmd:

```
./compound withdraw --opponent xxxx --asset xxxxxxx --amount 10000
```

### market
> Initiate a add market proposal

cmd:

```
//-s  symbol
//-a asset_id
//-c ctoken_asset_id
./compound add-market --s BTC --a xxxxxx --c yyyyyy
or
./compound am -s BTC -a xxxxx -c yyyyyyy
```

### close-market
> Initiate a closing market proposal

cmd:

```
./compound close-market --asset xxxxxxxx
or
./compound cm --asset xxxxxx
```

### open-market
> Initiate a opening market proposal

cmd:

```
./compound open-market --asset xxxxx
or
./compound om --asset xxxxxxx
```

### allowlist
> Initiate a allowlist proposal. 
> scope: temporarily only supports liquidation

cmd:

```
$compound allowlist add --user {user_id} --scope {scope}
$compound allowlist remove --user {user_id} --scope {scope}
```

### add-oracle-signer

cmd:

```
$compound add-oracle-signer --user xxx --key
```