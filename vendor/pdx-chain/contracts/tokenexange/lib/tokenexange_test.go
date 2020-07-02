/*************************************************************************
 * Copyright (C) 2016-2019 PDX Technologies, Inc. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @Time   : 2019-08-20 10:39
 * @Author : liangc
 *************************************************************************/

package tokenexangelib

import (
	"context"
	"fmt"
	"pdx-chain/accounts/abi/bind"
	"pdx-chain/common"
	"pdx-chain/ethclient"
	"pdx-chain/rpc"
	"testing"
)

func TestTokenList(t *testing.T) {
	ctx := context.Background()
	endpoint := "http://localhost:8545"
	exangeAddr := common.HexToAddress("0x123")

	client, err := rpc.Dial(endpoint)
	t.Log(err)
	ethClient := ethclient.NewClient(client)
	exange, err := NewExange(exangeAddr, ethClient)
	t.Log(err)
	opts := &bind.CallOpts{
		Context: ctx,
	}
	tokenlist, err := exange.TokenList(opts)
	for i, token := range tokenlist {
		ti, err := exange.TokenInfo(opts, token)
		t.Log(i, err, token.Hex(), ti)
	}
}

func TestABIS(t *testing.T) {
	fmt.Println(ExangeABI)
}
