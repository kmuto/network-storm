# ネットワーク検証ツール

- [ping](ping/README.md): ping応答ダミー `sim_ping.py`
- `snmpmetric`: SNMPメトリック応答ダミー

## 設定ファイルの作成
- `create_ip_and_delays.rb`: `sim_ping.py`や設定ファイル作成ツール向けのCSV生成。10.0.0.0ネットワーク決めうちにしている。生成したCSVファイルを`sim_ping.py`に指定する
  - 第1引数: 生成するCSVファイル名
  - 第2引数: IPアドレス数。省略すると1000
  - 第3引数: タイムアウトのホストの頻度。省略すると0.05 (5%)
  - 第4引数: 何番目から開始するか (※ホストを追加したくなったとき用)
- `create_mackerel_hosts.rb`: Mackerelホストの生成
  - 第1引数: IPアドレス一覧のCSVファイル名 (`create_ip_and_delays.rb`で生成したCSV)
  - 第2引数: 所属する「サービス:ロール」の文字列。省略すると`switches:tokyo`
  - `MACKEREL_APIKEY`環境変数付きで実行すること
- `retire_mackerel_hosts.rb`: Mackerelホストの退役。指定のサービス:ロールのホストを退役するので注意
  - 第1引数: サービス名。省略すると`swithces`
  - 第2引数: ロール名。省略すると`tokyo`
  - `MACKEREL_APIKEY`環境変数付きで実行すること
- `create_spconfig.rb`: pingメトリックツールの設定ファイルを生成
  - 第1引数: IPアドレス一覧のCSVファイル名 (`create_ip_and_delays.rb`で作成したCSV)
  - 第2引数: 書き出す設定ファイル名
  - `MACKEREL_APIKEY`環境変数付きで実行すること

