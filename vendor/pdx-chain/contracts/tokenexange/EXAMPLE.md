# Tokenexange 使用说明

## 接口与事件

> 接口符合 `web3` 的 `ABI` 规范，可以抽象成 `solidity contract`，
最简单的方式是使用 `web3.js` 直接用 `ABI + 合约地址` 来生成合约实例，适合 `H5` 端；
本例将提供 `web3j` 的调用方式，可以直接用在 `java` 后台和 `android` 端

合约地址: `0x0000000000000000000000000000000000000123`

代码与说明：

```
contract exange {

    /*
    列出当前链上符合 ERC20 接口规范的合约地址列表, 开发者可以根据列表来生成交易对
    例如：
        tokenlist 返回 [A,B,C,D]
        以 A 为交易区，我们可以展示 B/A 、C/A、D/A 三个交易对

        后面的接口参数中，我们会用 `address region` 表示交易区，`address target` 表示交易目标，
        交易对 `B/A` 等价于 `{region:A , target:B}`
    */
    function tokenList() public view returns (address[] memory);

    /*
    根据 ERC20 合约地址，获取摘要信息，可以用在展示交易区和交易对时使用
    */
    function tokenInfo(address) public view returns (string memory name, string memory symbol, uint256 totalSupply, uint8 decimals);

    /*
    挂买单
    参数说明：
        region : 交易区 erc20 合约地址
        target : 交易目标 erc20 合约地址
        regionAmount : 想要用 regionAmount 个 region token 换取 targetAmount 个 target token
        targetAmount : 想要用 regionAmount 个 region token 换取 targetAmount 个 target token

        由于合约不能直接处理浮点数，所以这里的兑换价格需要开发者自行计算,
        在精度相等时 price = regionAmount / targetAmount，否则需要先除精度
    */
    function bid(address region, address target, uint256 regionAmount, uint256 targetAmount) public payable;

    /*
    挂卖单
    参数说明：
        region : 交易区 erc20 合约地址
        target : 交易目标 erc20 合约地址
        regionAmount : 想要用 targetAmount 个 target token 换取 regionAmount 个 region token
        targetAmount : 想要用 targetAmount 个 target token 换取 regionAmount 个 region token
    */
    function ask(address region, address target, uint256 regionAmount, uint256 targetAmount) public payable;

    /*
    撤销订单(ask/bid)
        订单在达到 final 状态之前均可以撤销，撤销后状态变为 final 即不再参与撮合
    */
    function cancel(uint256 id) public payable;

    /*
    查询余额，根据 erc20 合约地址查询在 tokenexange 中的余额，
    余额是通过向 tokenexange 合约地址充值 region 对应的 token 得来的
    例如：
        用户 U 持有 A 资产，则 U 去执行 A.transfer(tokenexange.address,amount) 成功后，
        用户 U 再去执行 tokenexange.balanceOf(A) 时将会得到 (A.name,A,symbol,amount,decimals) 元组
    */
    function balanceOf(address region) public view returns (string memory name, string memory symbol, uint256 balance, uint8 decimals);

    /*
    余额提现，只要 balanceOf 能查询出来的余额都可以提现
    还以用户 U 和资产 A 来举例，提现操作相当于 A.transfer(U,tokenexange.balanceOf(A).balance)
    */
    function withdrawal(address region, uint256 amount) public payable;

    /*
    查询挂单列表，在指定的交易对上进行查询，只返回订单 ID 列表，
    注意不要修改列表顺序，列表已经按价格进行排序，ask 单是升序，bid 单是降序
    参数：
        orderType : 订单类型，可选值为 "bid" / "ask" 分别表示 买单 / 卖单
        region : 交易区 erc20 合约地址
        target : 交易目标 erc20 合约地址
    */
    function orderlist(string memory orderType, address region, address target) public view returns (uint256[] memory);

    /*
    获取订单详情
    详情包含了挂单时的全部信息，同时还包含了订单的当前状态，
    其中 regionComplete / targetComplete 是当前已经撮合成的数量
    isFinal == true 时表示订单为最终转改，不再参与撮合
    两种情况会让订单变为最终状态，一是撮合完成，二是撤单
    */
    function orderinfo(uint256 id) public view returns (
        string memory orderType,
        address region, address target,
        uint256 regionAmount, uint256 targetAmount,
        uint256 regionComplete, uint256 targetComplete,
        uint8 regionDecimals, uint8 targetDecimals,
        bool isFinal,
        address owner
    );
    /*
    查询我挂过的订单 (bid & ask)
    参数：
        addrs : 是一个数组，其长度必须是 1 或 3,
                是 1 时 addrs = [owner] 表示查询 owner 的全部挂单信息
                是 3 时 addrs = [owner,region,target] 表示查询 owner 在 target/region 交易区的挂单信息
        pageNum : 分页检索时用来表示页号，每页20条信息；
                TODO 目前没有实现分页，传 1 即可返回全部信息
    */
    function ownerOrder(address[] memory addrs, uint256 pageNum) public view returns (uint256[] memory);

    /*
    此事件记录订单状态变化，在撮合时触发，无论是否为最终状态都会触发
    记录被撮合的订单变化信息，主要包括如下属性
            ( orderid, owner, region, rc, regionAmount, target, tc, targetAmount )
      分别对应：
            ( 订单id, 订单创建人, 交易区, 操作(加/减), 交易区资产数量, 目标资产, 操作(加/减), 目标资产数量 )
      例如：
            (111,"0x1","0xA",+1,100,"0xB",-1,200)
            表示 111 这个订单成交信息为 A 资产增加 100, B 资产减少 200
            从这个资产变化甚至可以看出 111 是一个 ask 单，本次撮合卖出了 200B 收获了 100A
    */
    event Combination(uint256 indexed orderid, address indexed owner,
        address region, int8 rc, uint256 regionAmount,
        address target, int8 tc, uint256 targetAmount
    );
    // 此事件在挂买单成功时触发，用来通知 dapp 有新的买单产生
    event Bid(address indexed owner, uint256 indexed orderid);
    // 此事件在挂卖单成功时触发，用来通知 dapp 有新的卖单产生
    event Ask(address indexed owner, uint256 indexed orderid);
}
```

## 使用 web3j 调用合约

> 与 `web3.js` 一样，使用 `web3j` 调用合约时也可以直接根据 `ABI+合约地址` 构建出合约实例，
不同的是 `web3j` 需要先使用 `ABI` 生成合约的代理类，然后再用 `合约地址` 去实例化这个对象.

### web3j 安装

* windows/Linux 下载解压文件

```bash
wget https://github.com/web3j/web3j/releases/download/v4.4.0/web3j-4.4.0.zip
unzip web3j-4.4.0.zip
cd web3j-4.4.0/bin/
```
其中 `web3j` 为 `linux` 可执行文件 , `web3j.bat` 为 `windowx` 可执行文件

* mac

```bash
brew tap web3j/web3j
brew install web3j
```

### 用 ABI 生成 Java 类

* 将以下abi保存到exchange.abi文件

```json
[{"constant":false,"inputs":[{"name":"region","type":"address"},{"name":"target","type":"address"},{"name":"regionAmount","type":"uint256"},{"name":"targetAmount","type":"uint256"}],"name":"ask","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"region","type":"address"},{"name":"target","type":"address"},{"name":"regionAmount","type":"uint256"},{"name":"targetAmount","type":"uint256"}],"name":"bid","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"id","type":"uint256"}],"name":"cancel","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"name":"region","type":"address"},{"name":"amount","type":"uint256"}],"name":"withdrawal","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"orderid","type":"uint256"},{"indexed":true,"name":"owner","type":"address"},{"indexed":false,"name":"region","type":"address"},{"indexed":false,"name":"rc","type":"int8"},{"indexed":false,"name":"regionAmount","type":"uint256"},{"indexed":false,"name":"target","type":"address"},{"indexed":false,"name":"tc","type":"int8"},{"indexed":false,"name":"targetAmount","type":"uint256"}],"name":"Combination","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"orderid","type":"uint256"}],"name":"Bid","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"orderid","type":"uint256"}],"name":"Ask","type":"event"},{"constant":true,"inputs":[{"name":"region","type":"address"}],"name":"balanceOf","outputs":[{"name":"name","type":"string"},{"name":"symbol","type":"string"},{"name":"balance","type":"uint256"},{"name":"decimals","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"id","type":"uint256"}],"name":"orderinfo","outputs":[{"name":"orderType","type":"string"},{"name":"region","type":"address"},{"name":"target","type":"address"},{"name":"regionAmount","type":"uint256"},{"name":"targetAmount","type":"uint256"},{"name":"regionComplete","type":"uint256"},{"name":"targetComplete","type":"uint256"},{"name":"regionDecimals","type":"uint8"},{"name":"targetDecimals","type":"uint8"},{"name":"isFinal","type":"bool"},{"name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"orderType","type":"string"},{"name":"region","type":"address"},{"name":"target","type":"address"}],"name":"orderlist","outputs":[{"name":"","type":"uint256[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"addrs","type":"address[]"},{"name":"pageNum","type":"uint256"}],"name":"ownerOrder","outputs":[{"name":"","type":"uint256[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"tokenInfo","outputs":[{"name":"name","type":"string"},{"name":"symbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"tokenList","outputs":[{"name":"","type":"address[]"}],"payable":false,"stateMutability":"view","type":"function"}]
```

* 生成空的exchange.code文件

* 执行命令

```bash
web3j solidity generate --solidityTypes {path}/exchange.code {path}/exchange.abi -o {outpath} -p {package}
```


### 测试合约方法

```java
/**
 * 以tokenList为例 
 */
public class Test {

    private static Logger logger = LoggerFactory.getLogger(SpringbootApplicationTests.class);
    private static final String URL = "http://10.0.0.76:8545";
    private static final String CONTRACT_ADDRESS = "0x0000000000000000000000000000000000000123";
    private static final String PK = "f7b9a83c1a182fcddac14f90f686d4a08e8e43f1b617b9056f457c1837c9cb70";

    public void contextLoads() {
    }

    @Test
    public void tokenList() throws Exception {
        Web3j web3j = Web3j.build(new HttpService(URL));
        Exchange exchange = Exchange.load(CONTRACT_ADDRESS, web3j, Credentials.create(PK), new BigInteger("18000000000"), new BigInteger("5000000"));
        List<Address> addresses = exchange.tokenList().send().getValue();
        addresses.stream().forEach(address -> {
            logger.info("address==>{}", address.getValue());
        });
    }

}
```