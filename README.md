# 2025_2_Yabloko

Проект - Delivery Club

gRPC Generate:
```bash
protoc --go_out=. --go-grpc_out=. pkg/proto/embedding/embedding.proto
```

before deploy:
```bash
chmod +x monitoring/alertmanager/render_cfg.sh
./monitoring/alertmanager/render_cfg.sh
```
