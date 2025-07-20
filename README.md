# 启动方法
`gorder/`
- `docker compose up -d`
- `stripe listen --forward-to localhost:8284/api/webhook`

依次cd进入`gorder/internal/stock`, `gorder/internal/order`, `gorder/internal/payment`, 执行`air .`

# 项目笔记
要明确区分业务报错和系统报错!例子:
```go
func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) (res []*entity.Item, err error) {
	lockKey := getLockKey(query)
	if err = lock(ctx, lockKey); err != nil {
		return nil, errors.Wrapf(err, "redis lock error||key=%s", lockKey)
	}
	defer func() {
		f := logrus.Fields{
			"query": query,
			"res":   res,
		}
		if err != nil {
			logging.Errorf(ctx, f, "checkIfItemsInStock error||err=%v", err)
		} else {
			logging.Infof(ctx, f, "checkIfItemsInStock success")
		}

		if err = unlock(ctx, lockKey); err != nil {
			logging.Warnf(ctx, nil, "redis unlock error||err=%v", err)
		}
		fmt.Printf(">> defer %v\n", err) // nil
	}()

	for _, it := range query.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, it.ID)
		if err != nil {
			logging.Warnf(ctx, nil, "GetPriceByProductID error, item ID = %s, err = %v", it.ID, err)
			return nil, err
		}

		res = append(res, &entity.Item{
			ID:       it.ID,
			Quantity: it.Quantity,
			PriceID:  priceID,
		})
	}

	if err := h.checkStock(ctx, query.Items); err != nil {
		fmt.Printf(">> checkStock %v\n", err) // not enough
		return nil, err
	}
	return res, nil
}
```
以上代码在返回时会覆盖业务报错