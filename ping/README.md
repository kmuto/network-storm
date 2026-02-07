# pingレスポンスダミー

ホストの裏にあたかもネットワークがあるかのように、そのネットワーク範囲に届いたpingリクエストに対して任意の反応速度で返す。

## 必要なもの
- Python 3
- python3-scapy
- root権限

## delays.csv書式
```
ip,median_ms,jitter_ms,loss_rate
10.0.0.1,10,2,0
10.0.0.2,142,134,0.29
10.0.0.3,2201,500,0.93
```

- `ip`: IPアドレス
- `median_ms`: pingレスポンスの中央値（ms）
- `jitter_ms`: `median_ms`に対するランダムな振れ幅（ms）
- `loss_rate`: パケットロス率

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
