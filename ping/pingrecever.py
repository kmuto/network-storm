from scapy.all import *
import time
import threading

# 送信元Macアドレスを取得（dum0のもの、またはダミー）
MY_MAC = get_if_hwaddr("dum0")

def send_reply(pkt, delay):
    if delay > 0:
        time.sleep(delay)
    
    # 受信パケットのEthernet層から、逆向きのパケットを作る
    # dst は pingを打った側のMAC (pkt[Ether].src)
    # src は dum0のMAC (MY_MAC)
    reply = Ether(dst=pkt[Ether].src, src=MY_MAC) / \
            IP(src=pkt[IP].dst, dst=pkt[IP].src) / \
            ICMP(type=0, id=pkt[ICMP].id, seq=pkt[ICMP].seq)

    # ペイロード（データ部分）があれば引き継ぐ
    if pkt.haslayer(Raw):
        reply = reply / pkt[Raw].load

    # チェックサムの再計算を強制（Scapyは del すると送信時に再計算する）
    del reply[IP].chksum
    del reply[ICMP].chksum
    
    sendp(reply, iface="dum0", verbose=False)
    # デバッグ用に送信元/宛先MACを表示
    print(f"Reply: {pkt[IP].dst} -> {pkt[IP].src} (MAC: {MY_MAC} -> {pkt[Ether].src})")

def handle_packet(pkt):
    # ICMP Echo Request (type 8) かつ 自分から出たパケットを対象にする
    if ICMP in pkt and pkt[ICMP].type == 8:
        target_ip = pkt[IP].dst
        # IPの末尾によって遅延を変えるロジック（例：末尾 * 1ms）
        last_octet = int(target_ip.split('.')[-1])
        delay = last_octet * 0.001

        threading.Thread(target=send_reply, args=(pkt, delay)).start()

print("Monitoring dum0 for local ping traffic...")
# L2層でsniffを開始
sniff(iface="dum0", prn=handle_packet, filter="icmp", store=0)
