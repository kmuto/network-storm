# pingレスポンスダミー

pingレスポンスをさせるホストの裏にあたかもネットワークがあるかのように、そのネットワーク範囲に届いたpingリクエストに対して任意の反応速度で返す。

## delays.csv書式
```
ip,median_ms,jitter_ms
10.0.0.1,10,2
10.0.0.2,50,10
10.0.0.3,-1,0
```

- `ip`: IPアドレス
- `median_ms`: pingレスポンスの中央値（ms）
- `jitter_ms`: `median_ms`に対するランダムな振れ幅（ms）

median_msを負値にした場合はタイムアウトとして何も返さない。

## 実行書式
```
sudo python3 sim_ping.py --iface インターフェイス --net ネットワーク/マスク --csv CSVファイル
```

- `--iface`: ping送信側に対して待ち構えているインターフェイス名
- `--net`: ダミーとして構成するネットワークとネットマスク値
- `--csv`: CSVファイル

たとえば

```
sudo python3 sim_ping.py --iface enp0s3 --net 10.0.0.0/22 --csv delays.csv
```

## pingリクエスト側ホストの設定
ネットワーク・ネットマスク値の範囲に対して、sim_ping.pyを実行しているホストに向けるようルーティングする。

たとえば

```
sudo route add -net 10.0.0.0/22 gw ホストIPアドレス
```
