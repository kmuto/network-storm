# ネットワーク検証ツール

- [ping](ping/README.md): ping応答ダミー `sim_ping.py`
- [snmpstorm](snmpstorm/README.md): SNMPメトリック応答ダミー `snmpstorm`

## ping・snmpstormで共存するための設定
実際のインターフェイスは作らず、ダミーインターフェイスで返答するようにする。

以下は10.0.0.0/20ネットワークを作った例。

```
sudo ip link add snmp-dummy type dummy
sudo ip link set snmp-dummy up
sudo ip addr add 10.0.0.0/20 dev snmp-dummy
sudo ip route add local 10.0.0.0/20 dev snmp-dummy
sudo sysctl -w net.ipv4.icmp_echo_ignore_all=1
```

## 設定ファイルの作成
- `create_ip_and_delays.rb`: `sim_ping.py`や設定ファイル作成ツール向けのCSV生成。生成したCSVファイルを`sim_ping.py`に指定する
  - 第1引数: 生成するCSVファイル名
  - 第2引数: IPアドレス数。省略すると1000
  - 第3引数: 先頭IPアドレス2オクテット。省略すると10.0
  - 第4引数: 何番目から開始するか
- `create_mackerel_hosts.rb`: Mackerelホストの生成
  - 第1引数: IPアドレス一覧のCSVファイル名 (`create_ip_and_delays.rb`で生成したCSV)
  - 第2引数: 所属する「サービス:ロール」の文字列。省略すると`switches:tokyo`
  - `MACKEREL_APIKEY`環境変数付きで実行すること
- `retire_mackerel_hosts.rb`: Mackerelホストの退役。指定のサービス:ロールのホストを退役するので注意
  - 第1引数: サービス名。省略すると`switches`
  - 第2引数: ロール名。省略すると`tokyo`
  - `MACKEREL_APIKEY`環境変数付きで実行すること
- `create_spconfig.rb`: pingメトリックツールの設定ファイルを生成
  - 第1引数: IPアドレス一覧のCSVファイル名 (`create_ip_and_delays.rb`で作成したCSV)
  - 第2引数: 書き出す設定ファイル名
  - `MACKEREL_APIKEY`環境変数付きで実行すること
- `create_snmpconfig.rb`: snmpstormツールの設定ファイルを生成
  - 第1引数: IPアドレス一覧のCSVファイル名 (`create_ip_and_delays.rb`で作成したCSV)
  - 第2引数: 書き出す設定ファイル名
- `create_sabatrafficdconfig.rb`: sabatrafficdの設定ファイルを生成
  - 第1引数: 書き出す設定ファイル名
  - 第2引数: サービス名。省略すると`switches`
  - 第3引数: ロール名。省略すると`tokyo`
  - `MACKEREL_APIKEY`環境変数付きで実行すること

## 応答側の例
- 10.0.0.0/20ネットワークを仮想利用
- `create_ip_and_delays.rb`で`ip.csv`を作成
- `create_snmpconfig.rb`で`snmpstorm.csv`を作成
- 物理インターフェイスはenp0s3

```
sudo python3 sim_ping.py --iface enp0s3 --net 10.0.0.0/20 --csv ip.csv
./snmpstorm -csv snmpstorm.csv -iface enp0s3
```

## 送信側の例
