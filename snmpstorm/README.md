# SNMPレスポンスダミー

ホストの裏にあたかもネットワークがあるかのように、そのネットワーク範囲に届いたSNMPリクエストに対して適当な値を返す。

## 必要なもの
- Goビルド環境

## snmp_config.csv書式
```
ip,if_count,delay_ms
127.0.0.1,5,100
127.0.0.2,10,200
127.0.0.3,2,-1
```

- `ip`: IPアドレス
- `if_count`: ポート数
- `delay_ms`: 応答までの遅延 (ms)。負値の場合は返答しない

## 実行書式
```
./snmpstorm --iface インターフェイス --csv CSVファイル
```

- `--iface`: SNMPリクエスト側に対して待ち構えているインターフェイス名
- `--csv`: CSVファイル

たとえば

```
./snmpstorm --iface enp0s3 --csv snmp_config.csv
```
