# CMP DEMO



```bash 
$> cd lib/multi-party-sig-cmp && go mod download # multi-party-sig-cmp 依赖安装
$> cd ../.. && go mod download # cmp依赖安装
```

在文件夹中已经生成了三个钱包`w1.kgc,w2.kgc,w3.kgc`，可以直接使用。也可以通过`create`来生成新的钱包。

## 命令

首先需要创建一个消息转发服务，打开一个新的终端，运行`go run main.go serve`，在过程中不关闭该终端。

接下来打开多个终端，分别代表不同的钱包。

- 创建钱包
  - go run main.go create -w <wallet_path> -i <party_id> -s <parties> -t <threshold>
  - 比如 `go run main.go create -w w1.kgc -i 1 -s 1,2,3 -t 3`
- 转账
  - go run main.go transfer -w <wallet_path> -d <dist_address>
  - 比如 `go run main.go -w w1.kgc` ，会有一个默认目标地址,`0x921B004dc386ba15604bB97205Bb20988192DEDf`