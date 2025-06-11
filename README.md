# 启动方法
`gorder/`
- `docker compose up -d`
- `stripe listen --forward-to localhost:8284/api/webhook`

依次cd进入`gorder/internal/stock`, `gorder/internal/order`, `gorder/internal/payment`, 执行`air .`