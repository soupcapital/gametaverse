
# 获取token
    
    {"operationName":"tokens","variables":{"value":"ETH","id":"ETH"},"query":"query tokens($value: String, $id: String) {\n  asSymbol: tokens(where: {symbol: $value}, orderBy: totalLiquidity, orderDirection: desc) {\n    id\n    symbol\n    name\n    totalLiquidity\n    __typename\n  }\n  }\n"}

# 获取pair

    {"operationName":"pairs","variables":{"tokens":["0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"],"id":"ET"},"query":"query pairs($tokens: [Bytes]!, $id: String) {\n  as0: pairs(where: {token0_in: $tokens}) {\n    id\n token0Price\n token1Price\n   token0 {\n      id\n      symbol\n      name\n      __typename\n    }\n    token1 {\n      id\n      symbol\n      name\n      __typename\n    }\n    __typename\n  }\n  as1: pairs(where: {token1_in: $tokens}) {\n    id\n token0Price\n token1Price\n    token0 {\n      id\n      symbol\n      name\n      __typename\n    }\n    token1 {\n      id\n      symbol\n      name\n      __typename\n    }\n    __typename\n  }\n  }\n"}