package tests

import (
	"context"
	"fmt"
	"log"
	"testing"

	sw "github.com/peiyouyao/gorder/common/client/order"
	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	ctx    = context.Background()
	server = fmt.Sprintf("http://%s/api", viper.GetString("order.http-addr"))
	client *sw.ClientWithResponses
)

func TestMain(m *testing.M) {
	before()
}

func before() {
	log.Printf("server=%s", server)
	c, err := sw.NewClientWithResponses(server)
	if err != nil {
		log.Fatal(err)
	}
	client = c
}

func TestCreateOrder_success(t *testing.T) {
	customerID := "123"
	rsp, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerID,
		sw.PostCustomerCustomerIdOrdersJSONRequestBody{
			CustomerId: customerID,
			Items: []sw.ItemWithQuantity{
				{
					Id:       "prod_SSGOnM6DXikQ7y",
					Quantity: 1,
				},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	t.Logf("json200=%+v", rsp.JSON200)
	assert.Equal(t, 200, rsp.StatusCode())
	assert.Equal(t, 0, rsp.JSON200.Errno)
}

func TestCreateOrder_invalidParam(t *testing.T) {
	customerID := "123"
	rsp, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerID,
		sw.PostCustomerCustomerIdOrdersJSONRequestBody{
			CustomerId: customerID,
			Items:      nil,
		},
	)
	if err != nil {
		t.Error(err)
	}

	t.Logf("json200=%+v", rsp.JSON200)
	assert.Equal(t, 200, rsp.StatusCode())
	assert.Equal(t, 2, rsp.JSON200.Errno)
}
