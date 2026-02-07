import argparse
import csv
import threading
import time
import random
from scapy.all import *

def get_args():
    parser = argparse.ArgumentParser(description='Ping Response Simulator')
    parser.add_argument('--iface', required=True, help='物理インターフェイス名 (例: enp0s3)')
    parser.add_argument('--net', required=True, help='対象ネットワーク (例: 10.0.0.0/22)')
    parser.add_argument('--csv', required=True, help='遅延設定CSVファイル')
    return parser.parse_args()

def load_delays(file_path):
    delays = {}
    try:
        with open(file_path, mode='r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                delays[row['ip']] = {
                    'median': float(row['median_ms']),
                    'jitter': float(row['jitter_ms'])
                }
    except Exception as e:
        print(f"Error loading CSV: {e}")
    return delays

def send_reply(pkt, iface, my_mac, delay_ms):
    # 遅延実行 (ms -> sec)
    if delay_ms > 0:
        time.sleep(delay_ms / 1000.0)

    # 応答パケット作成
    reply = Ether(dst=pkt[Ether].src, src=my_mac) / \
            IP(src=pkt[IP].dst, dst=pkt[IP].src) / \
            ICMP(type=0, id=pkt[ICMP].id, seq=pkt[ICMP].seq)

    if pkt.haslayer(Raw):
        reply = reply / pkt[Raw].load

    sendp(reply, iface=iface, verbose=False)
    print(f"  [Reply] {pkt[IP].dst} -> {pkt[IP].src} ({delay_ms:.2f}ms)")

def handle_packet(pkt, iface, my_ip, my_mac, delay_settings):
    if ICMP in pkt and pkt[ICMP].type == 8:
        target_ip = pkt[IP].dst

        # 自分のIPからのパケットは無視（ループ防止）
        if pkt[IP].src == my_ip:
            return

        # 遅延設定の取得
        conf = delay_settings.get(target_ip)
        if conf:
            median = conf['median']
            jitter = conf['jitter']

            # 遅延が負の場合は応答しない
            if median < 0:
                print(f"  [Skip] {target_ip} (Configured as drop)")
                return

            # 中央値 ± 揺らぎ幅 でランダム計算
            actual_delay = median + random.uniform(-jitter, jitter)
            actual_delay = max(0, actual_delay) # 0以下にならないよう補正

            threading.Thread(target=send_reply, args=(pkt, iface, my_mac, actual_delay)).start()
        else:
            # CSVにないIPは即時応答（または無視ならreturn）
            threading.Thread(target=send_reply, args=(pkt, iface, my_mac, 0)).start()

def main():
    args = get_args()

    # 自分の情報を自動取得
    try:
        my_ip = get_if_addr(args.iface)
        my_mac = get_if_hwaddr(args.iface)
    except:
        print(f"インターフェイス {args.iface} が見つかりません。")
        return

    print(f"--- Simulator Started ---")
    print(f"Interface: {args.iface} ({my_ip} / {my_mac})")
    print(f"Target Net: {args.net}")

    delay_settings = load_delays(args.csv)
    print(f"Loaded {len(delay_settings)} IP settings from {args.csv}")

    # Sniff開始
    sniff_filter = f"icmp and dst net {args.net}"
    sniff(iface=args.iface, filter=sniff_filter,
          prn=lambda pkt: handle_packet(pkt, args.iface, my_ip, my_mac, delay_settings),
          store=0)

if __name__ == "__main__":
    main()
