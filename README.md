# 2025_2_Yabloko

Проект - Delivery Club

gRPC Generate:
```bash
protoc --go_out=. --go-grpc_out=. pkg/proto/embedding/embedding.proto
```

# Embedding Service Model:

Test Model in repo: MiniLM-L6-v2.Q8_0.gguf

Using Prod Model:
```bash
wget -O MiniLM-L12-118M-v2-Q8_0.gguf \
  "https://huggingface.co/mykor/paraphrase-multilingual-MiniLM-L12-v2.gguf/resolve/main/paraphrase-multilingual-MiniLM-L12-118M-v2-Q8_0.gguf?download=true"
```
Small Model:
```bash
sudo wget -O all-MiniLM-L6-v2-Q5_K_M.gguf \
  "https://huggingface.co/second-state/All-MiniLM-L6-v2-Embedding-GGUF/resolve/main/all-MiniLM-L6-v2-Q5_K_M.gguf?download=true"
```