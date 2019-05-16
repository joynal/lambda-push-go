# Push service

### Install packages
```
dep init
```

### Run test
```
go test
```

### Deploy

Development:

```
sls deploy -v --force --stage dev
```


Production:

```
sls deploy -v --force --stage prod
```
