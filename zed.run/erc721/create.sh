#!/usr/bin/env bash

graph init \
  --product hosted-service \
  --from-contract 0xA5F1Ea7DF861952863dF2e8d1312f7305dabf215 \
  --network matic \
  --abi ./IERC721Upgradeable.abi \
  cz-theng/zed-run-721