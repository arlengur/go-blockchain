### Service launch

```bash
go run simple/*.go
```

The server will start at http://localhost:8080

### Add block
```bash
curl -X POST http://localhost:8080/addblock \
     -H "Content-Type: application/json" \
     -d '{"value": 10}'  
```

### Get balance
```bash
curl -X POST http://localhost:8080/balance \
     -H "Content-Type: application/json" \
     -d '{"Address": "f479fb1daa4459080441de161324e94e94d8876e3e31d2b50ba6b2ae7fb05f7a"}'
```