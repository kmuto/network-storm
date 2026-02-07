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
    parser.add_argument('--csv', required=True, help='遅延・ロス設定CSVファイル')
    return parser.parse_args()

def load_delays(file_path):
    delays = {}
    try:
        with open(file_path, mode='r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                # loss_rateカラムがない場合は、medianが負なら1.0(100%ロス)とする
                l_rate = float(row.get('loss_rate', 0.0))
                m_val = float(row['median_ms'])
                
                if m_val < 0:
                    l_rate = 1.0

                delays[row['ip']] = {
                    'median': m_val,
                    'jitter': float(row['jitter_ms']),
                    'loss_rate': l_rate
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

        if pkt[IP].src == my_ip:
            return

        conf = delay_settings.get(target_ip)
        if conf:
            loss_rate = conf['loss_rate']
            
            # ロス判定: 0.0〜1.0の乱数がloss_rateを下回ったらドロップ
            if random.random() < loss_rate:
                print(f"  [Drop]  {target_ip} (Loss rate: {loss_rate*100:.1f}%)")
                return

            median = conf['median']
            jitter = conf['jitter']

            # 中央値 ± 揺らぎ幅 でランダム計算
            actual_delay = median + random.uniform(-jitter, jitter)
            actual_delay = max(0, actual_delay)

            threading.Thread(target=send_reply, args=(pkt, iface, my_mac, actual_delay)).start()
        else:
            threading.Thread(target=send_reply, args=(pkt, iface, my_mac, 0)).start()

def main():
    args = get_args()

    try:
        my_ip = get_if_addr(args.iface)
        my_mac = get_if_hwaddr(args.iface)
    except:
        print(f"インターフェイス {args.iface} が見つかりません。")
        return

    print(f"--- Simulator Started (Loss Rate Support) ---")
    print(f"Interface: {args.iface} ({my_ip} / {my_mac})")
    print(f"Target Net: {args.net}")

    delay_settings = load_delays(args.csv)
    print(f"Loaded {len(delay_settings)} IP settings from {args.csv}")

    sniff_filter = f"icmp and dst net {args.net}"
    sniff(iface=args.iface, filter=sniff_filter,
          prn=lambda pkt: handle_packet(pkt, args.iface, my_ip, my_mac, delay_settings),
          store=0)

if __name__ == "__main__":
    main()
