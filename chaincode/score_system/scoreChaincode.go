package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

// ScoreChaincode 积分|代币 链码结构体
type ScoreChaincode struct {
	contractapi.Contract
}

// TokenUnit 代币数据类型，后期可改为整形int64或者float64
type TokenUnit float64

// 货币类型标识
const (
	CNY int = iota
	USD
	EUR
	YEN
	GBP
)

// CurrencyType 交易货币类型
var CurrencyType = map[int]string{
	CNY: "CNY",
	USD: "USD",
	EUR: "EUR",
	YEN: "YEN",
	GBP: "GBP",
}

// Wallet 钱包
type Wallet struct {
	Address     string        `json:"address"`
	Balance     float64       `json:"balance"`
	DefaultUnit string        `json:"default_unit"`
	Exchange    string        `json:"exchange"` //交易所
	ServiceFee  TokenUnit     `json:"service_fee"`
	TotalCost   TokenUnit     `json:"total_cost"`
	WalletToken []WalletToken `json:"wallet_token"`
}

// WalletToken 钱包积分代币
type WalletToken struct {
	Token            Token     `json:"token"`
	AccumulatedToken TokenUnit `json:"accumulated_token"`
}

// Token 积分or代币
type Token struct {
	//Owner    string  `json:"owner"` //Wallet.address
	Amount   float64 `json:"amount"`
	Merchant string  `json:"merchant"`
}

// InitLedger 初始化示例账本
func (sc *ScoreChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	wallets := []Wallet{
		{
			Address:     "abc",
			Balance:     0,
			DefaultUnit: CurrencyType[CNY],
			Exchange:    "Charles",
			ServiceFee:  0,
			TotalCost:   0,
			WalletToken: []WalletToken{
				{
					Token:            Token{0, "apple"},
					AccumulatedToken: 0,
				},
			},
		},
		{
			Address:     "bcd",
			Balance:     0,
			DefaultUnit: CurrencyType[CNY],
			Exchange:    "Charles",
			ServiceFee:  0,
			TotalCost:   0,
			WalletToken: []WalletToken{
				{
					Token:            Token{0, "huawei"},
					AccumulatedToken: 0,
				},
			},
		},
	}
	for _, wallet := range wallets {
		walletByte, err := json.Marshal(wallet)
		err = ctx.GetStub().PutState(wallet.Address, walletByte)
		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

// QueryWallet 查询钱包
//{"Args": ["QueryWallet", "address"]}
func (sc *ScoreChaincode) QueryWallet(ctx contractapi.TransactionContextInterface, address string) (string, error) {
	walletbyte, err := ctx.GetStub().GetState(address)
	if err != nil {
		return "", fmt.Errorf("get state error: %s", err.Error())
	}
	fmt.Println(string(walletbyte))
	return string(walletbyte), nil
}

// DeleteWallet 删除钱包
//{"Args": ["DeleteWallet", "address"]}
func (sc *ScoreChaincode) DeleteWallet(ctx contractapi.TransactionContextInterface, address string) error {
	wallet, err := sc.QueryWallet(ctx, address)
	if wallet == "" {
		return fmt.Errorf("wallet: %s does not exist", address)
	}
	err = ctx.GetStub().DelState(address)
	if err != nil {
		return fmt.Errorf("delete wallet: %s , error", address)
	}
	fmt.Println("delete wallet success")
	return nil
}

// GetToken 获得积分|代币
func (sc *ScoreChaincode) GetToken(ctx contractapi.TransactionContextInterface, consumer string, merchant string, amountStr string) error {
	var flag = false
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return errors.New("amount is error,please check again")
	}
	resbyte, err := ctx.GetStub().GetState(consumer)
	if err != nil {
		return err
	}
	if resbyte == nil {
		fmt.Println("the wallet address does not exist, please create a new wallet first!")
		return errors.New("please create a new wallet")
	}
	wallet := Wallet{}
	err = json.Unmarshal(resbyte, &wallet)
	if err != nil {
		return errors.New("wallet unmarshal failed")
	}
	for i := 0; i < len(wallet.WalletToken); i++ {
		if wallet.WalletToken[i].Token.Merchant == merchant {
			wallet.WalletToken[i].Token.Amount += amount
			wallet.WalletToken[i].AccumulatedToken += TokenUnit(amount)
			flag = true
			break
		}
	}
	if !flag {
		fmt.Println("does not exist merchant token!")
		walletToken := WalletToken{
			Token: Token{
				Amount:   amount,
				Merchant: merchant,
			},
			AccumulatedToken: TokenUnit(amount),
		}
		// err = createToken(wallet, walletToken)
		wallet.WalletToken = append(wallet.WalletToken, walletToken)
	}
	res, _ := json.Marshal(wallet)

	err = ctx.GetStub().PutState(consumer, res)
	return nil
}

// CreateWallet 创建钱包
//{"Args": ["CreateWallet"]}
func (sc *ScoreChaincode) CreateWallet(ctx contractapi.TransactionContextInterface, address string) error {
	// address := newAddress()
	wallets := []Wallet{
		{
			Address:     address,
			Balance:     0,
			DefaultUnit: CurrencyType[CNY],
			Exchange:    "Charles",
			ServiceFee:  0,
			TotalCost:   0,
			WalletToken: []WalletToken{
				{
					Token:            Token{0, ""},
					AccumulatedToken: 0,
				},
			},
		},
	}
	for _, wallet := range wallets {
		walletByte, err := json.Marshal(wallet)
		err = ctx.GetStub().PutState(address, walletByte)
		if err != nil {
			return fmt.Errorf("error putstate: %s", err.Error())
		}
	}
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(ScoreChaincode))
	if err != nil {
		fmt.Printf("Error create score system chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting score system chaincode: %s", err.Error())
	}
}
