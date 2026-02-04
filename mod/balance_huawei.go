package mod

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	bss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/model"
	bssregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bss/v2/region"
)

// QueryHuaweiBalance queries Huawei Cloud account balance via BSS API.
func QueryHuaweiBalance(accessKey string, secretKey string, region string) (string, string, error) {
	if accessKey == "" || secretKey == "" {
		return "", "", fmt.Errorf("missing huaweicloud access key or secret")
	}
	if region == "" {
		region = "cn-north-4"
	}

	cred := basic.NewCredentialsBuilder().WithAk(accessKey).WithSk(secretKey).Build()
	client := bss.NewBssClient(bss.BssClientBuilder().WithCredential(cred).WithRegion(bssregion.ValueOf(region)).Build())

	request := &model.ShowCustomerAccountBalancesRequest{}
	response, err := client.ShowCustomerAccountBalances(request)
	if err != nil {
		return "", "", err
	}
	if response == nil || response.AccountBalances == nil || len(*response.AccountBalances) == 0 {
		return "", "", fmt.Errorf("empty huaweicloud balance response")
	}

	balance := (*response.AccountBalances)[0]
	amount := fmt.Sprintf("%.2f", balance.Amount)
	currency := balance.Currency
	if currency == "" {
		currency = "CNY"
	}
	return amount, currency, nil
}
